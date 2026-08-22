// LunarFishing.cs
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text.Json;
using System.Text.Json.Serialization;

class Config
{
    [JsonPropertyName("locations")]
    public Dictionary<string, Location> Locations { get; set; } = new Dictionary<string, Location>();
}

class Location
{
    [JsonPropertyName("lat")]
    public double Lat { get; set; }
    [JsonPropertyName("lon")]
    public double Lon { get; set; }
}

class MoonData
{
    public string Phase { get; set; }
    public double Illumination { get; set; }
    public double Age { get; set; }
    public string Zodiac { get; set; }
}

class SolunarData
{
    public double Major1Start { get; set; }
    public double Major1End { get; set; }
    public double Major2Start { get; set; }
    public double Major2End { get; set; }
    public double Minor1Start { get; set; }
    public double Minor1End { get; set; }
    public double Minor2Start { get; set; }
    public double Minor2End { get; set; }
    public double MoonTransit { get; set; }
}

class LunarFishing
{
    private const double DEFAULT_LAT = 0.0;
    private const double DEFAULT_LON = 0.0;
    private const string CONFIG_FILE = "lunar_fishing_config.json";
    private static readonly JsonSerializerOptions Options = new JsonSerializerOptions { WriteIndented = true };

    static double JulianDay(DateTime dt)
    {
        int year = dt.Year;
        int month = dt.Month;
        double day = dt.Day + dt.Hour/24.0 + dt.Minute/1440.0 + dt.Second/86400.0;
        if (month <= 2) { year--; month += 12; }
        int A = year / 100;
        int B = 2 - A + A / 4;
        return (int)(365.25 * (year + 4716)) + (int)(30.6001 * (month + 1)) + day + B - 1524.5;
    }

    static (double lon, double lat) MoonPosition(double jd)
    {
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

        double lon = L_prime + (6.289 * Math.Sin(M_prime) + 1.274 * Math.Sin(2*D - M_prime) + 0.658 * Math.Sin(2*D) + 0.214 * Math.Sin(2*M_prime) - 0.186 * Math.Sin(M) - 0.114 * Math.Sin(2*F)) * Math.PI / 180;
        double lat = (5.128 * Math.Sin(F) + 0.280 * Math.Sin(M_prime + F) + 0.278 * Math.Sin(M_prime - F) + 0.173 * Math.Sin(2*D - F)) * Math.PI / 180;
        return (lon, lat);
    }

    static double SunPosition(double jd)
    {
        double T = (jd - 2451545.0) / 36525.0;
        double M = ((357.5291 + 35999.0503 * T) % 360) * Math.PI / 180;
        double C = 1.9146 * Math.Sin(M) + 0.0200 * Math.Sin(2*M) + 0.0003 * Math.Sin(3*M);
        double lon = ((280.4665 + 36000.7698 * T + C) % 360) * Math.PI / 180;
        return lon;
    }

    static double GreenwichSiderealTime(double jd)
    {
        double T = (jd - 2451545.0) / 36525.0;
        double gmst = 280.46061837 + 360.98564736629 * (jd - 2451545.0) + 0.000387933 * T*T - T*T*T / 38710000.0;
        gmst = ((gmst % 360) + 360) % 360;
        return gmst / 15.0;
    }

    static MoonData MoonPhase(DateTime dt)
    {
        double jd = JulianDay(dt);
        var (lonMoon, _) = MoonPosition(jd);
        double lonSun = SunPosition(jd);

        double elong = lonMoon - lonSun;
        elong = Math.Atan2(Math.Sin(elong), Math.Cos(elong));
        double phaseAngle = Math.Atan2(Math.Sin(elong), Math.Cos(elong));
        double illumination = (1 + Math.Cos(phaseAngle)) / 2;

        double age = (jd - 2451550.1) / 29.53058867;
        age = ((age % 29.53058867) + 29.53058867) % 29.53058867;

        string phase;
        if (age < 1.0) phase = "New Moon";
        else if (age < 7.38) phase = "Waxing Crescent";
        else if (age < 8.38) phase = "First Quarter";
        else if (age < 14.77) phase = "Waxing Gibbous";
        else if (age < 15.77) phase = "Full Moon";
        else if (age < 22.15) phase = "Waning Gibbous";
        else if (age < 23.15) phase = "Last Quarter";
        else phase = "Waning Crescent";

        string[] signs = {"Aries","Taurus","Gemini","Cancer","Leo","Virgo","Libra","Scorpio","Sagittarius","Capricorn","Aquarius","Pisces"};
        double lonDeg = ((lonMoon * 180 / Math.PI) % 360 + 360) % 360;
        int idx = (int)(lonDeg / 30);
        string zodiac = signs[idx];

        return new MoonData { Phase = phase, Illumination = illumination * 100, Age = age, Zodiac = zodiac };
    }

    static SolunarData Solunar(DateTime dt, double lat, double lon)
    {
        double jd = JulianDay(dt);
        var (lonMoon, _) = MoonPosition(jd);
        double lonMoonDeg = ((lonMoon * 180 / Math.PI) % 360 + 360) % 360;
        double gst = GreenwichSiderealTime(jd);
        double lst = ((gst + lon / 15.0) % 24.0 + 24.0) % 24.0;
        double moonRA = lonMoonDeg / 15.0;
        double moonTransit = ((moonRA - lst) % 24.0 + 24.0) % 24.0;

        return new SolunarData
        {
            Major1Start = moonTransit - 1.0,
            Major1End = moonTransit + 1.0,
            Major2Start = moonTransit + 12.0 - 1.0,
            Major2End = moonTransit + 12.0 + 1.0,
            Minor1Start = moonTransit + 6.0 - 0.5,
            Minor1End = moonTransit + 6.0 + 0.5,
            Minor2Start = moonTransit + 18.0 - 0.5,
            Minor2End = moonTransit + 18.0 + 0.5,
            MoonTransit = moonTransit
        };
    }

    static int FishingRating(string phase)
    {
        switch (phase)
        {
            case "Full Moon": case "New Moon": return 4;
            case "Waxing Gibbous": case "Waning Gibbous": case "First Quarter": case "Last Quarter": return 3;
            case "Waxing Crescent": case "Waning Crescent": return 2;
            default: return 1;
        }
    }

    static string FormatTime(double hours)
    {
        int h = (int)hours;
        int m = (int)((hours - h) * 60);
        return $"{h:D2}:{m:D2}";
    }

    static void Render(DateTime dt, double lat, double lon, bool phaseOnly)
    {
        var moonData = MoonPhase(dt);
        var solunarData = Solunar(dt, lat, lon);
        int rating = FishingRating(moonData.Phase);

        if (phaseOnly)
        {
            Console.WriteLine(moonData.Phase);
            return;
        }

        string[] ratings = {"Poor", "Fair", "Good", "Excellent"};
        string[] ratingEmoji = {"⭐", "⭐⭐", "⭐⭐⭐", "⭐⭐⭐⭐"};

        Console.WriteLine("\n🎣 Lunar Fishing Calendar");
        string latStr = $"{Math.Abs(lat):F2}°{(lat >= 0 ? 'N' : 'S')}";
        string lonStr = $"{Math.Abs(lon):F2}°{(lon >= 0 ? 'E' : 'W')}";
        Console.WriteLine($"Location: {latStr}, {lonStr}");
        Console.WriteLine($"Date: {dt:yyyy-MM-dd HH:mm}");

        Console.WriteLine($"\n🌓 Moon Phase: {moonData.Phase} ({moonData.Illumination:F1}% illuminated)");
        Console.WriteLine($"🌙 Moon Age: {moonData.Age:F1} days");
        Console.WriteLine($"♑ Zodiac: {moonData.Zodiac}");

        Console.WriteLine("\n🎯 Solunar Feeding Periods:");
        var periods = new[]
        {
            new { Start = solunarData.Major1Start, End = solunarData.Major1End, Label = "Major Period 1", Emoji = ratingEmoji[rating-1], Rating = ratings[rating-1] },
            new { Start = solunarData.Major2Start, End = solunarData.Major2End, Label = "Major Period 2", Emoji = ratingEmoji[rating-1], Rating = ratings[rating-1] },
            new { Start = solunarData.Minor1Start, End = solunarData.Minor1End, Label = "Minor Period 1", Emoji = "⭐", Rating = "Good" },
            new { Start = solunarData.Minor2Start, End = solunarData.Minor2End, Label = "Minor Period 2", Emoji = "⭐", Rating = "Good" },
        };
        foreach (var p in periods)
        {
            if (p.Start >= 0 && p.Start < 24)
                Console.WriteLine($"  {p.Label}: {FormatTime(p.Start)} – {FormatTime(p.End)} ({p.Emoji} {p.Rating})");
        }

        Console.WriteLine($"\n⭐ Best Fishing Rating: {ratings[rating-1]} ({ratingEmoji[rating-1]})");

        // Sunrise/sunset (simplified)
        int dayOfYear = dt.DayOfYear;
        double declination = 23.44 * Math.Sin((284 + dayOfYear) * 360 * Math.PI / 180 / 365);
        double latRad = lat * Math.PI / 180;
        double decRad = declination * Math.PI / 180;
        double cosHA = -Math.Tan(latRad) * Math.Tan(decRad);
        double ha;
        if (cosHA < -1) ha = Math.PI;
        else if (cosHA > 1) ha = 0;
        else ha = Math.Acos(cosHA);
        double dayLength = ha * 2 / (Math.PI / 12);
        double noon = 12.0 - lon / 15.0;
        double sunrise = noon - dayLength / 2;
        double sunset = noon + dayLength / 2;
        Console.WriteLine($"🌅 Sunrise: {FormatTime(sunrise)} | 🌇 Sunset: {FormatTime(sunset)}");

        double moonRise = (solunarData.MoonTransit + 6) % 24;
        double moonSet = (solunarData.MoonTransit + 18) % 24;
        Console.WriteLine($"🌙 Moonrise: {FormatTime(moonRise)} | 🌇 Moonset: {FormatTime(moonSet)}");
    }

    static void Main(string[] args)
    {
        var parsed = ParseArgs(args);

        Config config = new Config();
        if (File.Exists(CONFIG_FILE))
        {
            string json = File.ReadAllText(CONFIG_FILE);
            config = JsonSerializer.Deserialize<Config>(json) ?? new Config();
        }

        if (parsed.ContainsKey("list-locations"))
        {
            if (config.Locations.Count == 0)
            {
                Console.WriteLine("No saved locations.");
                return;
            }
            Console.WriteLine("\n📍 Saved Locations:");
            foreach (var kv in config.Locations)
            {
                string latStr = $"{Math.Abs(kv.Value.Lat):F2}°{(kv.Value.Lat >= 0 ? 'N' : 'S')}";
                string lonStr = $"{Math.Abs(kv.Value.Lon):F2}°{(kv.Value.Lon >= 0 ? 'E' : 'W')}";
                Console.WriteLine($"  {kv.Key}: {latStr}, {lonStr}");
            }
            return;
        }

        double lat = parsed.ContainsKey("lat") ? double.Parse(parsed["lat"]) : DEFAULT_LAT;
        double lon = parsed.ContainsKey("lon") ? double.Parse(parsed["lon"]) : DEFAULT_LON;

        if (parsed.ContainsKey("save-location"))
        {
            config.Locations[parsed["save-location"]] = new Location { Lat = lat, Lon = lon };
            string json = JsonSerializer.Serialize(config, Options);
            File.WriteAllText(CONFIG_FILE, json);
            Console.WriteLine($"✅ Location '{parsed["save-location"]}' saved.");
        }

        if (parsed.ContainsKey("use-location"))
        {
            string name = parsed["use-location"];
            if (config.Locations.ContainsKey(name))
            {
                var loc = config.Locations[name];
                lat = loc.Lat;
                lon = loc.Lon;
                Console.WriteLine($"📍 Using saved location: {name}");
            }
            else
            {
                Console.WriteLine($"Location '{name}' not found.");
                return;
            }
        }

        DateTime dt = DateTime.UtcNow;
        if (parsed.ContainsKey("date"))
        {
            dt = DateTime.Parse(parsed["date"] + " 00:00:00").ToUniversalTime();
        }

        Render(dt, lat, lon, parsed.ContainsKey("phase-only"));
    }

    static Dictionary<string, string> ParseArgs(string[] args)
    {
        var dict = new Dictionary<string, string>();
        for (int i=0; i<args.Length; i++)
        {
            if (args[i].StartsWith("--"))
            {
                string key = args[i].Substring(2);
                if (i+1 < args.Length && !args[i+1].StartsWith("--"))
                    dict[key] = args[++i];
                else
                    dict[key] = "";
            }
        }
        return dict;
    }
}
