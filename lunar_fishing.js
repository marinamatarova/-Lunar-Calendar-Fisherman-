// lunar_fishing.js
#!/usr/bin/env node
const fs = require('fs');
const { program } = require('commander');

const DEFAULT_LAT = 0.0;
const DEFAULT_LON = 0.0;
const CONFIG_FILE = 'lunar_fishing_config.json';

function julianDay(date) {
    let year = date.getFullYear();
    let month = date.getMonth() + 1;
    let day = date.getDate() + date.getHours()/24 + date.getMinutes()/1440 + date.getSeconds()/86400;
    if (month <= 2) { year--; month += 12; }
    let A = Math.floor(year / 100);
    let B = 2 - A + Math.floor(A / 4);
    return Math.floor(365.25 * (year + 4716)) + Math.floor(30.6001 * (month + 1)) + day + B - 1524.5;
}

function moonPosition(jd) {
    let T = (jd - 2451545.0) / 36525.0;
    let L_prime = 218.3165 + 481267.8813 * T;
    let D = 297.8502 + 445267.1114 * T;
    let M = 357.5291 + 35999.0503 * T;
    let M_prime = 134.9634 + 477198.8676 * T;
    let F = 93.2720 + 483202.0175 * T;

    L_prime = ((L_prime % 360) + 360) % 360 * Math.PI / 180;
    D = ((D % 360) + 360) % 360 * Math.PI / 180;
    M = ((M % 360) + 360) % 360 * Math.PI / 180;
    M_prime = ((M_prime % 360) + 360) % 360 * Math.PI / 180;
    F = ((F % 360) + 360) % 360 * Math.PI / 180;

    let lon = L_prime + (6.289 * Math.sin(M_prime) + 1.274 * Math.sin(2*D - M_prime) + 0.658 * Math.sin(2*D) + 0.214 * Math.sin(2*M_prime) - 0.186 * Math.sin(M) - 0.114 * Math.sin(2*F)) * Math.PI / 180;
    let lat = (5.128 * Math.sin(F) + 0.280 * Math.sin(M_prime + F) + 0.278 * Math.sin(M_prime - F) + 0.173 * Math.sin(2*D - F)) * Math.PI / 180;
    return { lon, lat };
}

function sunPosition(jd) {
    let T = (jd - 2451545.0) / 36525.0;
    let M = ((357.5291 + 35999.0503 * T) % 360) * Math.PI / 180;
    let C = 1.9146 * Math.sin(M) + 0.0200 * Math.sin(2*M) + 0.0003 * Math.sin(3*M);
    let lon = ((280.4665 + 36000.7698 * T + C) % 360) * Math.PI / 180;
    return lon;
}

function greenwichSiderealTime(jd) {
    let T = (jd - 2451545.0) / 36525.0;
    let gmst = 280.46061837 + 360.98564736629 * (jd - 2451545.0) + 0.000387933 * T*T - T*T*T / 38710000.0;
    gmst = ((gmst % 360) + 360) % 360;
    return gmst / 15.0;
}

function moonPhase(date) {
    let jd = julianDay(date);
    let { lon: lonMoon } = moonPosition(jd);
    let lonSun = sunPosition(jd);

    let elong = lonMoon - lonSun;
    elong = Math.atan2(Math.sin(elong), Math.cos(elong));
    let phaseAngle = Math.atan2(Math.sin(elong), Math.cos(elong));
    let illumination = (1 + Math.cos(phaseAngle)) / 2;

    let age = (jd - 2451550.1) / 29.53058867;
    age = ((age % 29.53058867) + 29.53058867) % 29.53058867;

    let phase;
    if (age < 1.0) phase = "New Moon";
    else if (age < 7.38) phase = "Waxing Crescent";
    else if (age < 8.38) phase = "First Quarter";
    else if (age < 14.77) phase = "Waxing Gibbous";
    else if (age < 15.77) phase = "Full Moon";
    else if (age < 22.15) phase = "Waning Gibbous";
    else if (age < 23.15) phase = "Last Quarter";
    else phase = "Waning Crescent";

    const signs = ["Aries","Taurus","Gemini","Cancer","Leo","Virgo","Libra","Scorpio","Sagittarius","Capricorn","Aquarius","Pisces"];
    let lonDeg = ((lonMoon * 180 / Math.PI) % 360 + 360) % 360;
    let idx = Math.floor(lonDeg / 30);
    let zodiac = signs[idx];

    return { phase, illumination: illumination * 100, age, zodiac, lonMoon, jd };
}

function solunar(date, lat, lon) {
    let jd = julianDay(date);
    let { lon: lonMoon } = moonPosition(jd);
    let lonMoonDeg = ((lonMoon * 180 / Math.PI) % 360 + 360) % 360;
    let gst = greenwichSiderealTime(jd);
    let lst = ((gst + lon / 15.0) % 24.0 + 24.0) % 24.0;
    let moonRA = lonMoonDeg / 15.0;
    let moonTransit = ((moonRA - lst) % 24.0 + 24.0) % 24.0;
    let major1Start = moonTransit - 1.0;
    let major1End = moonTransit + 1.0;
    let major2Start = moonTransit + 12.0 - 1.0;
    let major2End = moonTransit + 12.0 + 1.0;
    let minor1Start = moonTransit + 6.0 - 0.5;
    let minor1End = moonTransit + 6.0 + 0.5;
    let minor2Start = moonTransit + 18.0 - 0.5;
    let minor2End = moonTransit + 18.0 + 0.5;
    return { major1Start, major1End, major2Start, major2End, minor1Start, minor1End, minor2Start, minor2End, moonTransit };
}

function fishingRating(phase) {
    switch (phase) {
        case "Full Moon": case "New Moon": return 4;
        case "Waxing Gibbous": case "Waning Gibbous": case "First Quarter": case "Last Quarter": return 3;
        case "Waxing Crescent": case "Waning Crescent": return 2;
        default: return 1;
    }
}

function formatTime(hours) {
    const h = Math.floor(hours);
    const m = Math.floor((hours - h) * 60);
    return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`;
}

function render(date, lat, lon, phaseOnly) {
    let moonData = moonPhase(date);
    let solunarData = solunar(date, lat, lon);
    let rating = fishingRating(moonData.phase);

    if (phaseOnly) {
        console.log(moonData.phase);
        return;
    }

    const ratings = ["Poor", "Fair", "Good", "Excellent"];
    const ratingEmoji = ["⭐", "⭐⭐", "⭐⭐⭐", "⭐⭐⭐⭐"];

    console.log(`\n🎣 Lunar Fishing Calendar`);
    const latStr = `${Math.abs(lat).toFixed(2)}°${lat >= 0 ? 'N' : 'S'}`;
    const lonStr = `${Math.abs(lon).toFixed(2)}°${lon >= 0 ? 'E' : 'W'}`;
    console.log(`Location: ${latStr}, ${lonStr}`);
    console.log(`Date: ${date.toISOString().slice(0,16).replace('T',' ')}`);

    console.log(`\n🌓 Moon Phase: ${moonData.phase} (${moonData.illumination.toFixed(1)}% illuminated)`);
    console.log(`🌙 Moon Age: ${moonData.age.toFixed(1)} days`);
    console.log(`♑ Zodiac: ${moonData.zodiac}`);

    console.log(`\n🎯 Solunar Feeding Periods:`);
    const periods = [
        { start: solunarData.major1Start, end: solunarData.major1End, label: 'Major Period 1', emoji: ratingEmoji[rating-1], rating: ratings[rating-1] },
        { start: solunarData.major2Start, end: solunarData.major2End, label: 'Major Period 2', emoji: ratingEmoji[rating-1], rating: ratings[rating-1] },
        { start: solunarData.minor1Start, end: solunarData.minor1End, label: 'Minor Period 1', emoji: '⭐', rating: 'Good' },
        { start: solunarData.minor2Start, end: solunarData.minor2End, label: 'Minor Period 2', emoji: '⭐', rating: 'Good' },
    ];
    for (const p of periods) {
        if (p.start >= 0 && p.start < 24) {
            console.log(`  ${p.label}: ${formatTime(p.start)} – ${formatTime(p.end)} (${p.emoji} ${p.rating})`);
        }
    }

    console.log(`\n⭐ Best Fishing Rating: ${ratings[rating-1]} (${ratingEmoji[rating-1]})`);

    // Sunrise/sunset (simplified)
    const dayOfYear = Math.floor((date - new Date(date.getFullYear(), 0, 0)) / (1000*60*60*24));
    const declination = 23.44 * Math.sin((284 + dayOfYear) * 360 * Math.PI / 180 / 365);
    const latRad = lat * Math.PI / 180;
    const decRad = declination * Math.PI / 180;
    let cosHA = -Math.tan(latRad) * Math.tan(decRad);
    let ha;
    if (cosHA < -1) ha = Math.PI;
    else if (cosHA > 1) ha = 0;
    else ha = Math.acos(cosHA);
    const dayLength = ha * 2 / (Math.PI / 12);
    const noon = 12.0 - lon / 15.0;
    const sunrise = noon - dayLength / 2;
    const sunset = noon + dayLength / 2;
    console.log(`🌅 Sunrise: ${formatTime(sunrise)} | 🌇 Sunset: ${formatTime(sunset)}`);

    const moonRise = (solunarData.moonTransit + 6) % 24;
    const moonSet = (solunarData.moonTransit + 18) % 24;
    console.log(`🌙 Moonrise: ${formatTime(moonRise)} | 🌇 Moonset: ${formatTime(moonSet)}`);
}

program
    .option('--date <date>', 'YYYY-MM-DD')
    .option('--lat <lat>', 'Latitude (positive North)', parseFloat, DEFAULT_LAT)
    .option('--lon <lon>', 'Longitude (positive East)', parseFloat, DEFAULT_LON)
    .option('--phase-only', 'Output only phase name')
    .option('--save-location <name>', 'Save location with name')
    .option('--use-location <name>', 'Use saved location')
    .option('--list-locations', 'List saved locations')
    .parse(process.argv);

const opts = program.opts();

// Load config
let config = { locations: {} };
if (fs.existsSync(CONFIG_FILE)) {
    try {
        config = JSON.parse(fs.readFileSync(CONFIG_FILE));
    } catch (e) {}
}

if (opts.listLocations) {
    if (Object.keys(config.locations).length === 0) {
        console.log('No saved locations.');
        process.exit(0);
    }
    console.log('\n📍 Saved Locations:');
    for (const [name, loc] of Object.entries(config.locations)) {
        const latStr = `${Math.abs(loc.lat).toFixed(2)}°${loc.lat >= 0 ? 'N' : 'S'}`;
        const lonStr = `${Math.abs(loc.lon).toFixed(2)}°${loc.lon >= 0 ? 'E' : 'W'}`;
        console.log(`  ${name}: ${latStr}, ${lonStr}`);
    }
    process.exit(0);
}

if (opts.saveLocation) {
    config.locations[opts.saveLocation] = { lat: opts.lat, lon: opts.lon };
    fs.writeFileSync(CONFIG_FILE, JSON.stringify(config, null, 2));
    console.log(`✅ Location '${opts.saveLocation}' saved.`);
}

let lat = opts.lat || DEFAULT_LAT;
let lon = opts.lon || DEFAULT_LON;

if (opts.useLocation) {
    if (config.locations[opts.useLocation]) {
        lat = config.locations[opts.useLocation].lat;
        lon = config.locations[opts.useLocation].lon;
        console.log(`📍 Using saved location: ${opts.useLocation}`);
    } else {
        console.log(`Location '${opts.useLocation}' not found.`);
        process.exit(1);
    }
}

let dt = new Date();
if (opts.date) {
    dt = new Date(opts.date + 'T00:00:00');
}
render(dt, lat, lon, opts.phaseOnly);
