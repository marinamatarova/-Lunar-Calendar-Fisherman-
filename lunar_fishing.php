# lunar_fishing.php
#!/usr/bin/env php
<?php

define('DEFAULT_LAT', 0.0);
define('DEFAULT_LON', 0.0);
define('CONFIG_FILE', 'lunar_fishing_config.json');

function julianDay($dt) {
    $year = (int)$dt->format('Y');
    $month = (int)$dt->format('m');
    $day = (float)$dt->format('d') + (float)$dt->format('H')/24 + (float)$dt->format('i')/1440 + (float)$dt->format('s')/86400;
    if ($month <= 2) { $year--; $month += 12; }
    $A = (int)($year / 100);
    $B = 2 - $A + (int)($A / 4);
    return (int)(365.25 * ($year + 4716)) + (int)(30.6001 * ($month + 1)) + $day + $B - 1524.5;
}

function moonPosition($jd) {
    $T = ($jd - 2451545.0) / 36525.0;
    $L_prime = 218.3165 + 481267.8813 * $T;
    $D = 297.8502 + 445267.1114 * $T;
    $M = 357.5291 + 35999.0503 * $T;
    $M_prime = 134.9634 + 477198.8676 * $T;
    $F = 93.2720 + 483202.0175 * $T;

    $L_prime = fmod($L_prime, 360) * M_PI / 180;
    $D = fmod($D, 360) * M_PI / 180;
    $M = fmod($M, 360) * M_PI / 180;
    $M_prime = fmod($M_prime, 360) * M_PI / 180;
    $F = fmod($F, 360) * M_PI / 180;

    $lon = $L_prime + (6.289 * sin($M_prime) + 1.274 * sin(2*$D - $M_prime) + 0.658 * sin(2*$D) + 0.214 * sin(2*$M_prime) - 0.186 * sin($M) - 0.114 * sin(2*$F)) * M_PI / 180;
    $lat = (5.128 * sin($F) + 0.280 * sin($M_prime + $F) + 0.278 * sin($M_prime - $F) + 0.173 * sin(2*$D - $F)) * M_PI / 180;
    return [$lon, $lat];
}

function sunPosition($jd) {
    $T = ($jd - 2451545.0) / 36525.0;
    $M = fmod(357.5291 + 35999.0503 * $T, 360) * M_PI / 180;
    $C = 1.9146 * sin($M) + 0.0200 * sin(2*$M) + 0.0003 * sin(3*$M);
    $lon = fmod(280.4665 + 36000.7698 * $T + $C, 360) * M_PI / 180;
    return $lon;
}

function greenwichSiderealTime($jd) {
    $T = ($jd - 2451545.0) / 36525.0;
    $gmst = 280.46061837 + 360.98564736629 * ($jd - 2451545.0) + 0.000387933 * $T*$T - $T*$T*$T / 38710000.0;
    $gmst = fmod($gmst, 360.0);
    return $gmst / 15.0;
}

function moonPhase($dt) {
    $jd = julianDay($dt);
    list($lon_moon) = moonPosition($jd);
    $lon_sun = sunPosition($jd);

    $elong = $lon_moon - $lon_sun;
    $elong = atan2(sin($elong), cos($elong));
    $phase_angle = atan2(sin($elong), cos($elong));
    $illumination = (1 + cos($phase_angle)) / 2;

    $age = ($jd - 2451550.1) / 29.53058867;
    $age = fmod($age, 29.53058867);
    if ($age < 0) $age += 29.53058867;

    if ($age < 1.0) $phase = "New Moon";
    elseif ($age < 7.38) $phase = "Waxing Crescent";
    elseif ($age < 8.38) $phase = "First Quarter";
    elseif ($age < 14.77) $phase = "Waxing Gibbous";
    elseif ($age < 15.77) $phase = "Full Moon";
    elseif ($age < 22.15) $phase = "Waning Gibbous";
    elseif ($age < 23.15) $phase = "Last Quarter";
    else $phase = "Waning Crescent";

    $signs = ["Aries","Taurus","Gemini","Cancer","Leo","Virgo","Libra","Scorpio","Sagittarius","Capricorn","Aquarius","Pisces"];
    $lon_deg = fmod($lon_moon * 180 / M_PI, 360);
    if ($lon_deg < 0) $lon_deg += 360;
    $idx = (int)($lon_deg / 30);
    $zodiac = $signs[$idx];

    return ['phase' => $phase, 'illumination' => $illumination * 100, 'age' => $age, 'zodiac' => $zodiac, 'lon_moon' => $lon_moon, 'jd' => $jd];
}

function solunar($dt, $lat, $lon) {
    $jd = julianDay($dt);
    list($lon_moon) = moonPosition($jd);
    $lon_moon_deg = fmod($lon_moon * 180 / M_PI, 360);
    if ($lon_moon_deg < 0) $lon_moon_deg += 360;
    $gst = greenwichSiderealTime($jd);
    $lst = fmod($gst + $lon / 15.0, 24.0);
    $moon_ra = $lon_moon_deg / 15.0;
    $moon_transit = fmod($moon_ra - $lst, 24.0);
    if ($moon_transit < 0) $moon_transit += 24.0;
    return [
        'major1_start' => $moon_transit - 1.0,
        'major1_end' => $moon_transit + 1.0,
        'major2_start' => $moon_transit + 12.0 - 1.0,
        'major2_end' => $moon_transit + 12.0 + 1.0,
        'minor1_start' => $moon_transit + 6.0 - 0.5,
        'minor1_end' => $moon_transit + 6.0 + 0.5,
        'minor2_start' => $moon_transit + 18.0 - 0.5,
        'minor2_end' => $moon_transit + 18.0 + 0.5,
        'moon_transit' => $moon_transit
    ];
}

function fishingRating($phase) {
    switch ($phase) {
        case "Full Moon": case "New Moon": return 4;
        case "Waxing Gibbous": case "Waning Gibbous": case "First Quarter": case "Last Quarter": return 3;
        case "Waxing Crescent": case "Waning Crescent": return 2;
        default: return 1;
    }
}

function formatTime($hours) {
    $h = (int)$hours;
    $m = (int)(($hours - $h) * 60);
    return sprintf("%02d:%02d", $h, $m);
}

function render($dt, $lat, $lon, $phaseOnly) {
    $moonData = moonPhase($dt);
    $solunarData = solunar($dt, $lat, $lon);
    $rating = fishingRating($moonData['phase']);

    if ($phaseOnly) {
        echo $moonData['phase'] . "\n";
        return;
    }

    $ratings = ["Poor", "Fair", "Good", "Excellent"];
    $ratingEmoji = ["⭐", "⭐⭐", "⭐⭐⭐", "⭐⭐⭐⭐"];

    echo "\n🎣 Lunar Fishing Calendar\n";
    $latStr = abs($lat) . "°" . ($lat >= 0 ? 'N' : 'S');
    $lonStr = abs($lon) . "°" . ($lon >= 0 ? 'E' : 'W');
    echo "Location: $latStr, $lonStr\n";
    echo "Date: " . $dt->format('Y-m-d H:i') . "\n";

    echo "\n🌓 Moon Phase: {$moonData['phase']} (" . number_format($moonData['illumination'], 1) . "% illuminated)\n";
    echo "🌙 Moon Age: " . number_format($moonData['age'], 1) . " days\n";
    echo "♑ Zodiac: {$moonData['zodiac']}\n";

    echo "\n🎯 Solunar Feeding Periods:\n";
    $periods = [
        ['start' => $solunarData['major1_start'], 'end' => $solunarData['major1_end'], 'label' => 'Major Period 1', 'emoji' => $ratingEmoji[$rating-1], 'rating' => $ratings[$rating-1]],
        ['start' => $solunarData['major2_start'], 'end' => $solunarData['major2_end'], 'label' => 'Major Period 2', 'emoji' => $ratingEmoji[$rating-1], 'rating' => $ratings[$rating-1]],
        ['start' => $solunarData['minor1_start'], 'end' => $solunarData['minor1_end'], 'label' => 'Minor Period 1', 'emoji' => '⭐', 'rating' => 'Good'],
        ['start' => $solunarData['minor2_start'], 'end' => $solunarData['minor2_end'], 'label' => 'Minor Period 2', 'emoji' => '⭐', 'rating' => 'Good'],
    ];
    foreach ($periods as $p) {
        if ($p['start'] >= 0 && $p['start'] < 24) {
            echo "  {$p['label']}: " . formatTime($p['start']) . " – " . formatTime($p['end']) . " ({$p['emoji']} {$p['rating']})\n";
        }
    }

    echo "\n⭐ Best Fishing Rating: {$ratings[$rating-1]} ({$ratingEmoji[$rating-1]})\n";

    // Sunrise/sunset (simplified)
    $day_of_year = (int)$dt->format('z') + 1;
    $declination = 23.44 * sin((284 + $day_of_year) * 360 * M_PI / 180 / 365);
    $lat_rad = $lat * M_PI / 180;
    $dec_rad = $declination * M_PI / 180;
    $cos_ha = -tan($lat_rad) * tan($dec_rad);
    if ($cos_ha < -1) $ha = M_PI;
    elseif ($cos_ha > 1) $ha = 0;
    else $ha = acos($cos_ha);
    $day_length = $ha * 2 / (M_PI / 12);
    $noon = 12.0 - $lon / 15.0;
    $sunrise = $noon - $day_length / 2;
    $sunset = $noon + $day_length / 2;
    echo "🌅 Sunrise: " . formatTime($sunrise) . " | 🌇 Sunset: " . formatTime($sunset) . "\n";

    $moonRise = ($solunarData['moon_transit'] + 6) % 24;
    $moonSet = ($solunarData['moon_transit'] + 18) % 24;
    echo "🌙 Moonrise: " . formatTime($moonRise) . " | 🌇 Moonset: " . formatTime($moonSet) . "\n";
}

$opts = getopt("", ["date:", "lat:", "lon:", "phase-only", "save-location:", "use-location:", "list-locations"]);
$dateStr = $opts['date'] ?? null;
$lat = isset($opts['lat']) ? (float)$opts['lat'] : DEFAULT_LAT;
$lon = isset($opts['lon']) ? (float)$opts['lon'] : DEFAULT_LON;
$phaseOnly = isset($opts['phase-only']);
$saveLocation = $opts['save-location'] ?? null;
$useLocation = $opts['use-location'] ?? null;
$listLocations = isset($opts['list-locations']);

$config = ["locations" => []];
if (file_exists(CONFIG_FILE)) {
    $config = json_decode(file_get_contents(CONFIG_FILE), true) ?? $config;
}

if ($listLocations) {
    if (empty($config["locations"])) {
        echo "No saved locations.\n";
        exit(0);
    }
    echo "\n📍 Saved Locations:\n";
    foreach ($config["locations"] as $name => $loc) {
        $latStr = abs($loc['lat']) . "°" . ($loc['lat'] >= 0 ? 'N' : 'S');
        $lonStr = abs($loc['lon']) . "°" . ($loc['lon'] >= 0 ? 'E' : 'W');
        echo "  $name: $latStr, $lonStr\n";
    }
    exit(0);
}

if ($saveLocation) {
    $config["locations"][$saveLocation] = ["lat" => $lat, "lon" => $lon];
    file_put_contents(CONFIG_FILE, json_encode($config, JSON_PRETTY_PRINT));
    echo "✅ Location '$saveLocation' saved.\n";
}

if ($useLocation && isset($config["locations"][$useLocation])) {
    $lat = $config["locations"][$useLocation]["lat"];
    $lon = $config["locations"][$useLocation]["lon"];
    echo "📍 Using saved location: $useLocation\n";
} elseif ($useLocation) {
    echo "Location '$useLocation' not found.\n";
    exit(1);
}

$dt = new DateTime('now', new DateTimeZone('UTC'));
if ($dateStr) {
    $dt = new DateTime($dateStr . ' 00:00:00', new DateTimeZone('UTC'));
}

render($dt, $lat, $lon, $phaseOnly);
?>
