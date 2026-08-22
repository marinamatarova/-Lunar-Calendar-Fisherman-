// LunarFishing.java
import java.io.*;
import java.nio.file.*;
import java.time.*;
import java.time.format.*;
import java.util.*;
import com.google.gson.*;

class Config {
    public Map<String, Location> locations = new HashMap<>();
}

class Location {
    public double lat;
    public double lon;
}

class MoonData {
    String phase;
    double illumination;
    double age;
    String zodiac;
}

class SolunarData {
    double major1Start, major1End, major2Start, major2End;
    double minor1Start, minor1End, minor2Start, minor2End;
    double moonTransit;
}

public class LunarFishing {
    private static final double DEFAULT_LAT = 0.0;
    private static final double DEFAULT_LON = 0.0;
    private static final String CONFIG_FILE = "lunar_fishing_config.json";
    private static final Gson gson = new GsonBuilder().setPrettyPrinting().create();

    public static double julianDay(LocalDateTime dt) {
        int year = dt.getYear();
        int month = dt.getMonthValue();
        double day = dt.getDayOfMonth() + dt.getHour()/24.0 + dt.getMinute()/1440.0 + dt.getSecond()/86400.0;
        if (month <= 2) { year--; month += 12; }
        int A = year / 100;
        int B = 2 - A + A / 4;
        return (int)(365.25 * (year + 4716)) + (int)(30.6001 * (month + 1)) + day + B - 1524.5;
    }

    public static double[] moonPosition(double jd) {
        double T = (jd - 2451545.0) / 36525.0;
        double L_prime = 218.3165 + 481267.8813 * T;
        double D = 297.8502 + 445267.1114 * T;
        double M = 357.5291 + 35999.0503 * T;
        double M_prime = 134.9634 + 477198.8676 * T;
        double F = 93.2720 + 483202.0175 * T;

        L_prime = ((L_prime % 360) + 360) % 360 * Math.PI / 180;
        D = ((D % 360) + 360) % 360 * Math.PI / 180;
        M = ((M % 360) + 360) % 360 * Math.PI / 180;
        M_prime = ((M_prime % 360) + 360) % 360 * Math.PI / 180;
        F = ((F % 360) + 360) % 360 * Math.PI / 180;

        double lon = L_prime + (6.289 * Math.sin(M_prime) + 1.274 * Math.sin(2*D - M_prime) + 0.658 * Math.sin(2*D) + 0.214 * Math.sin(2*M_prime) - 0.186 * Math.sin(M) - 0.114 * Math.sin(2*F)) * Math.PI / 180;
        double lat = (5.128 * Math.sin(F) + 0.280 * Math.sin(M_prime + F) + 0.278 * Math.sin(M_prime - F) + 0.173 * Math.sin(2*D - F)) * Math.PI / 180;
        return new double[]{lon, lat};
    }

    public static double sunPosition(double jd) {
        double T = (jd - 2451545.0) / 36525.0;
        double M = ((357.5291 + 35999.0503 * T) % 360) * Math.PI / 180;
        double C = 1.9146 * Math.sin(M) + 0.0200 * Math.sin(2*M) + 0.0003 * Math.sin(3*M);
        double lon = ((280.4665 + 36000.7698 * T + C) % 360) * Math.PI / 180;
        return lon;
    }

    public static double greenwichSiderealTime(double jd) {
        double T = (jd - 2451545.0) / 36525.0;
        double gmst = 280.46061837 + 360.98564736629 * (jd - 2451545.0) + 0.000387933 * T*T - T*T*T / 38710000.0;
        gmst = ((gmst % 360) + 360) % 360;
        return gmst / 15.0;
    }

    public static MoonData moonPhase(LocalDateTime dt) {
        double jd = julianDay(dt);
        double[] moon = moonPosition(jd);
        double lonMoon = moon[0];
        double lonSun = sunPosition(jd);

        double elong = lonMoon - lonSun;
        elong = Math.atan2(Math.sin(elong), Math.cos(elong));
        double phaseAngle = Math.atan2(Math.sin(elong), Math.cos(elong));
        double illumination = (1 + Math.cos(phaseAngle)) / 2;

        double age = (jd - 2451550.1) / 29.53058867;
        age = ((age % 29.53058867) + 29.53058867) % 29.53058867;

        String phase;
        if (age < 1.0) phase = "New Moon";
        else if (age < 7.38) phase = "Waxing Crescent";
        else if (age < 8.38) phase = "First Quarter";
        else if (age < 14.77) phase = "Waxing Gibbous";
        else if (age < 15.77) phase = "Full Moon";
        else if (age < 22.15) phase = "Waning Gibbous";
        else if (age < 23.15) phase = "Last Quarter";
        else phase = "Waning Crescent";

        String[] signs = {"Aries","Taurus","Gemini","Cancer","Leo","Virgo","Libra","Scorpio","Sagittarius","Capricorn","Aquarius","Pisces"};
        double lonDeg = ((lonMoon * 180 / Math.PI) % 360 + 360) % 360;
        int idx = (int)(lonDeg / 30);
        String zodiac = signs[idx];

        MoonData data = new MoonData();
        data.phase = phase;
        data.illumination = illumination * 100;
        data.age = age;
        data.zodiac = zodiac;
        return data;
    }

    public static SolunarData solunar(LocalDateTime dt, double lat, double lon) {
        double jd = julianDay(dt);
        double[] moon = moonPosition(jd);
        double lonMoonDeg = ((moon[0] * 180 / Math.PI) % 360 + 360) % 360;
        double gst = greenwichSiderealTime(jd);
        double lst = ((gst + lon / 15.0) % 24.0 + 24.0) % 24.0;
        double moonRA = lonMoonDeg / 15.0;
        double moonTransit = ((moonRA - lst) % 24.0 + 24.0) % 24.0;

        SolunarData data = new SolunarData();
        data.major1Start = moonTransit - 1.0;
        data.major1End = moonTransit + 1.0;
        data.major2Start = moonTransit + 12.0 - 1.0;
        data.major2End = moonTransit + 12.0 + 1.0;
        data.minor1Start = moonTransit + 6.0 - 0.5;
        data.minor1End = moonTransit + 6.0 + 0.5;
        data.minor2Start = moonTransit + 18.0 - 0.5;
        data.minor2End = moonTransit + 18.0 + 0.5;
        data.moonTransit = moonTransit;
        return data;
    }

    public static int fishingRating(String phase) {
        switch (phase) {
            case "Full Moon": case "New Moon": return 4;
            case "Waxing Gibbous": case "Waning Gibbous": case "First Quarter": case "Last Quarter": return 3;
            case "Waxing Crescent": case "Waning Crescent": return 2;
            default: return 1;
        }
    }

    public static String formatTime(double hours) {
        int h = (int)hours;
        int m = (int)((hours - h) * 60);
        return String.format("%02d:%02d", h, m);
    }

    public static void render(LocalDateTime dt, double lat, double lon, boolean phaseOnly) throws Exception {
        MoonData moonData = moonPhase(dt);
        SolunarData solunarData = solunar(dt, lat, lon);
        int rating = fishingRating(moonData.phase);

        if (phaseOnly) {
            System.out.println(moonData.phase);
            return;
        }

        String[] ratings = {"Poor", "Fair", "Good", "Excellent"};
        String[] ratingEmoji = {"⭐", "⭐⭐", "⭐⭐⭐", "⭐⭐⭐⭐"};

        System.out.println("\n🎣 Lunar Fishing Calendar");
        String latStr = String.format("%.2f°%c", Math.abs(lat), lat >= 0 ? 'N' : 'S');
        String lonStr = String.format("%.2f°%c", Math.abs(lon), lon >= 0 ? 'E' : 'W');
        System.out.printf("Location: %s, %s\n", latStr, lonStr);
        System.out.printf("Date: %s\n", dt.format(DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm")));

        System.out.printf("\n🌓 Moon Phase: %s (%.1f%% illuminated)\n", moonData.phase, moonData.illumination);
        System.out.printf("🌙 Moon Age: %.1f days\n", moonData.age);
        System.out.printf("♑ Zodiac: %s\n", moonData.zodiac);

        System.out.println("\n🎯 Solunar Feeding Periods:");
        double[][] periods = {
            {solunarData.major1Start, solunarData.major1End, 0},
            {solunarData.major2Start, solunarData.major2End, 1},
            {solunarData.minor1Start, solunarData.minor1End, 2},
            {solunarData.minor2Start, solunarData.minor2End, 3},
        };
        String[] labels = {"Major Period 1", "Major Period 2", "Minor Period 1", "Minor Period 2"};
        for (int i = 0; i < periods.length; i++) {
            double start = periods[i][0];
            double end = periods[i][1];
            if (start >= 0 && start < 24) {
                String emoji = i < 2 ? ratingEmoji[rating-1] : "⭐";
                String rate = i < 2 ? ratings[rating-1] : "Good";
                System.out.printf("  %s: %s – %s (%s %s)\n", labels[i], formatTime(start), formatTime(end), emoji, rate);
            }
        }

        System.out.printf("\n⭐ Best Fishing Rating: %s (%s)\n", ratings[rating-1], ratingEmoji[rating-1]);

        // Sunrise/sunset (simplified)
        int dayOfYear = dt.getDayOfYear();
        double declination = 23.44 * Math.sin((284 + dayOfYear) * 360 * Math.PI / 180 / 365);
        double latRad = lat * Math.PI / 180;
        double decRad = declination * Math.PI / 180;
        double cosHA = -Math.tan(latRad) * Math.tan(decRad);
        double ha;
        if (cosHA < -1) ha = Math.PI;
        else if (cosHA > 1) ha = 0;
        else ha = Math.acos(cosHA);
        double dayLength = ha * 2 / (Math.PI / 12);
        double noon = 12.0 - lon / 15.0;
        double sunrise = noon - dayLength / 2;
        double sunset = noon + dayLength / 2;
        System.out.printf("🌅 Sunrise: %s | 🌇 Sunset: %s\n", formatTime(sunrise), formatTime(sunset));

        double moonRise = (solunarData.moonTransit + 6) % 24;
        double moonSet = (solunarData.moonTransit + 18) % 24;
        System.out.printf("🌙 Moonrise: %s | 🌇 Moonset: %s\n", formatTime(moonRise), formatTime(moonSet));
    }

    public static void main(String[] args) throws Exception {
        Map<String, String> params = new HashMap<>();
        for (int i=0; i<args.length; i++) {
            if (args[i].startsWith("--")) {
                String key = args[i].substring(2);
                if (i+1 < args.length && !args[i+1].startsWith("--")) {
                    params.put(key, args[++i]);
                } else {
                    params.put(key, "");
                }
            }
        }

        Config config = new Config();
        if (Files.exists(Paths.get(CONFIG_FILE))) {
            String json = new String(Files.readAllBytes(Paths.get(CONFIG_FILE)));
            config = gson.fromJson(json, Config.class);
        }

        if (params.containsKey("list-locations")) {
            if (config.locations.isEmpty()) {
                System.out.println("No saved locations.");
                return;
            }
            System.out.println("\n📍 Saved Locations:");
            for (Map.Entry<String, Location> entry : config.locations.entrySet()) {
                String name = entry.getKey();
                Location loc = entry.getValue();
                String latStr = String.format("%.2f°%c", Math.abs(loc.lat), loc.lat >= 0 ? 'N' : 'S');
                String lonStr = String.format("%.2f°%c", Math.abs(loc.lon), loc.lon >= 0 ? 'E' : 'W');
                System.out.printf("  %s: %s, %s\n", name, latStr, lonStr);
            }
            return;
        }

        double lat = params.containsKey("lat") ? Double.parseDouble(params.get("lat")) : DEFAULT_LAT;
        double lon = params.containsKey("lon") ? Double.parseDouble(params.get("lon")) : DEFAULT_LON;

        if (params.containsKey("save-location")) {
            Location loc = new Location();
            loc.lat = lat;
            loc.lon = lon;
            config.locations.put(params.get("save-location"), loc);
            Files.write(Paths.get(CONFIG_FILE), gson.toJson(config).getBytes());
            System.out.printf("✅ Location '%s' saved.\n", params.get("save-location"));
        }

        if (params.containsKey("use-location")) {
            String name = params.get("use-location");
            if (config.locations.containsKey(name)) {
                Location loc = config.locations.get(name);
                lat = loc.lat;
                lon = loc.lon;
                System.out.printf("📍 Using saved location: %s\n", name);
            } else {
                System.out.printf("Location '%s' not found.\n", name);
                return;
            }
        }

        LocalDateTime dt = LocalDateTime.now(ZoneOffset.UTC);
        if (params.containsKey("date")) {
            LocalDate date = LocalDate.parse(params.get("date"));
            dt = LocalDateTime.of(date, LocalTime.of(0, 0));
        }

        render(dt, lat, lon, params.containsKey("phase-only"));
    }
}
