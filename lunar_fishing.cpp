// lunar_fishing.cpp
#include <iostream>
#include <fstream>
#include <string>
#include <vector>
#include <map>
#include <cmath>
#include <ctime>
#include <iomanip>
#include <sstream>
#include <nlohmann/json.hpp>
#include <getopt.h>

using namespace std;
using json = nlohmann::json;

const double DEFAULT_LAT = 0.0;
const double DEFAULT_LON = 0.0;
const string CONFIG_FILE = "lunar_fishing_config.json";

struct Location {
    double lat, lon;
};

struct MoonData {
    string phase;
    double illumination;
    double age;
    string zodiac;
};

struct SolunarData {
    double major1Start, major1End, major2Start, major2End;
    double minor1Start, minor1End, minor2Start, minor2End;
    double moonTransit;
};

double julianDay(const tm& dt) {
    int year = dt.tm_year + 1900;
    int month = dt.tm_mon + 1;
    double day = dt.tm_mday + dt.tm_hour/24.0 + dt.tm_min/1440.0 + dt.tm_sec/86400.0;
    if (month <= 2) { year--; month += 12; }
    int A = year / 100;
    int B = 2 - A + A / 4;
    return (int)(365.25 * (year + 4716)) + (int)(30.6001 * (month + 1)) + day + B - 1524.5;
}

pair<double, double> moonPosition(double jd) {
    double T = (jd - 2451545.0) / 36525.0;
    double L_prime = 218.3165 + 481267.8813 * T;
    double D = 297.8502 + 445267.1114 * T;
    double M = 357.5291 + 35999.0503 * T;
    double M_prime = 134.9634 + 477198.8676 * T;
    double F = 93.2720 + 483202.0175 * T;

    L_prime = fmod(L_prime, 360) * M_PI / 180;
    D = fmod(D, 360) * M_PI / 180;
    M = fmod(M, 360) * M_PI / 180;
    M_prime = fmod(M_prime, 360) * M_PI / 180;
    F = fmod(F, 360) * M_PI / 180;

    double lon = L_prime + (6.289 * sin(M_prime) + 1.274 * sin(2*D - M_prime) + 0.658 * sin(2*D) + 0.214 * sin(2*M_prime) - 0.186 * sin(M) - 0.114 * sin(2*F)) * M_PI / 180;
    double lat = (5.128 * sin(F) + 0.280 * sin(M_prime + F) + 0.278 * sin(M_prime - F) + 0.173 * sin(2*D - F)) * M_PI / 180;
    return {lon, lat};
}

double sunPosition(double jd) {
    double T = (jd - 2451545.0) / 36525.0;
    double M = fmod(357.5291 + 35999.0503 * T, 360) * M_PI / 180;
    double C = 1.9146 * sin(M) + 0.0200 * sin(2*M) + 0.0003 * sin(3*M);
    double lon = fmod(280.4665 + 36000.7698 * T + C, 360) * M_PI / 180;
    return lon;
}

double greenwichSiderealTime(double jd) {
    double T = (jd - 2451545.0) / 36525.0;
    double gmst = 280.46061837 + 360.98564736629 * (jd - 2451545.0) + 0.000387933 * T*T - T*T*T / 38710000.0;
    gmst = fmod(gmst, 360.0);
    if (gmst < 0) gmst += 360.0;
    return gmst / 15.0;
}

MoonData moonPhase(const tm& dt) {
    double jd = julianDay(dt);
    auto [lonMoon, _] = moonPosition(jd);
    double lonSun = sunPosition(jd);

    double elong = lonMoon - lonSun;
    elong = atan2(sin(elong), cos(elong));
    double phaseAngle = atan2(sin(elong), cos(elong));
    double illumination = (1 + cos(phaseAngle)) / 2;

    double age = (jd - 2451550.1) / 29.53058867;
    age = fmod(age, 29.53058867);
    if (age < 0) age += 29.53058867;

    string phase;
    if (age < 1.0) phase = "New Moon";
    else if (age < 7.38) phase = "Waxing Crescent";
    else if (age < 8.38) phase = "First Quarter";
    else if (age < 14.77) phase = "Waxing Gibbous";
    else if (age < 15.77) phase = "Full Moon";
    else if (age < 22.15) phase = "Waning Gibbous";
    else if (age < 23.15) phase = "Last Quarter";
    else phase = "Waning Crescent";

    vector<string> signs = {"Aries","Taurus","Gemini","Cancer","Leo","Virgo","Libra","Scorpio","Sagittarius","Capricorn","Aquarius","Pisces"};
    double lonDeg = fmod(lonMoon * 180 / M_PI, 360);
    if (lonDeg < 0) lonDeg += 360;
    int idx = (int)(lonDeg / 30);
    string zodiac = signs[idx];

    return {phase, illumination * 100, age, zodiac};
}

SolunarData solunar(const tm& dt, double lat, double lon) {
    double jd = julianDay(dt);
    auto [lonMoon, _] = moonPosition(jd);
    double lonMoonDeg = fmod(lonMoon * 180 / M_PI, 360);
    if (lonMoonDeg < 0) lonMoonDeg += 360;
    double gst = greenwichSiderealTime(jd);
    double lst = fmod(gst + lon / 15.0, 24.0);
    if (lst < 0) lst += 24.0;
    double moonRA = lonMoonDeg / 15.0;
    double moonTransit = fmod(moonRA - lst, 24.0);
    if (moonTransit < 0) moonTransit += 24.0;

    return {
        moonTransit - 1.0, moonTransit + 1.0,
        moonTransit + 12.0 - 1.0, moonTransit + 12.0 + 1.0,
        moonTransit + 6.0 - 0.5, moonTransit + 6.0 + 0.5,
        moonTransit + 18.0 - 0.5, moonTransit + 18.0 + 0.5,
        moonTransit
    };
}

int fishingRating(const string& phase) {
    if (phase == "Full Moon" || phase == "New Moon") return 4;
    if (phase == "Waxing Gibbous" || phase == "Waning Gibbous" || phase == "First Quarter" || phase == "Last Quarter") return 3;
    if (phase == "Waxing Crescent" || phase == "Waning Crescent") return 2;
    return 1;
}

string formatTime(double hours) {
    int h = (int)hours;
    int m = (int)((hours - h) * 60);
    char buf[6];
    snprintf(buf, sizeof(buf), "%02d:%02d", h, m);
    return string(buf);
}

void render(const tm& dt, double lat, double lon, bool phaseOnly) {
    MoonData moonData = moonPhase(dt);
    SolunarData solunarData = solunar(dt, lat, lon);
    int rating = fishingRating(moonData.phase);

    if (phaseOnly) {
        cout << moonData.phase << "\n";
        return;
    }

    vector<string> ratings = {"Poor", "Fair", "Good", "Excellent"};
    vector<string> ratingEmoji = {"⭐", "⭐⭐", "⭐⭐⭐", "⭐⭐⭐⭐"};

    cout << "\n🎣 Lunar Fishing Calendar\n";
    string latStr = to_string(abs(lat)).substr(0,5) + "°" + (lat >= 0 ? "N" : "S");
    string lonStr = to_string(abs(lon)).substr(0,5) + "°" + (lon >= 0 ? "E" : "W");
    cout << "Location: " << latStr << ", " << lonStr << "\n";
    char dateBuf[20];
    strftime(dateBuf, sizeof(dateBuf), "%Y-%m-%d %H:%M", &dt);
    cout << "Date: " << dateBuf << "\n";

    cout << "\n🌓 Moon Phase: " << moonData.phase << " (" << fixed << setprecision(1) << moonData.illumination << "% illuminated)\n";
    cout << "🌙 Moon Age: " << fixed << setprecision(1) << moonData.age << " days\n";
    cout << "♑ Zodiac: " << moonData.zodiac << "\n";

    cout << "\n🎯 Solunar Feeding Periods:\n";
    struct Period { double start, end; string label; string emoji; string rating; };
    vector<Period> periods = {
        {solunarData.major1Start, solunarData.major1End, "Major Period 1", ratingEmoji[rating-1], ratings[rating-1]},
        {solunarData.major2Start, solunarData.major2End, "Major Period 2", ratingEmoji[rating-1], ratings[rating-1]},
        {solunarData.minor1Start, solunarData.minor1End, "Minor Period 1", "⭐", "Good"},
        {solunarData.minor2Start, solunarData.minor2End, "Minor Period 2", "⭐", "Good"},
    };
    for (auto& p : periods) {
        if (p.start >= 0 && p.start < 24) {
            cout << "  " << p.label << ": " << formatTime(p.start) << " – " << formatTime(p.end) << " (" << p.emoji << " " << p.rating << ")\n";
        }
    }

    cout << "\n⭐ Best Fishing Rating: " << ratings[rating-1] << " (" << ratingEmoji[rating-1] << ")\n";

    // Sunrise/sunset (simplified)
    int dayOfYear = dt.tm_yday + 1;
    double declination = 23.44 * sin((284 + dayOfYear) * 360 * M_PI / 180 / 365);
    double latRad = lat * M_PI / 180;
    double decRad = declination * M_PI / 180;
    double cosHA = -tan(latRad) * tan(decRad);
    double ha;
    if (cosHA < -1) ha = M_PI;
    else if (cosHA > 1) ha = 0;
    else ha = acos(cosHA);
    double dayLength = ha * 2 / (M_PI / 12);
    double noon = 12.0 - lon / 15.0;
    double sunrise = noon - dayLength / 2;
    double sunset = noon + dayLength / 2;
    cout << "🌅 Sunrise: " << formatTime(sunrise) << " | 🌇 Sunset: " << formatTime(sunset) << "\n";

    double moonRise = fmod(solunarData.moonTransit + 6, 24.0);
    double moonSet = fmod(solunarData.moonTransit + 18, 24.0);
    cout << "🌙 Moonrise: " << formatTime(moonRise) << " | 🌇 Moonset: " << formatTime(moonSet) << "\n";
}

int main(int argc, char* argv[]) {
    static struct option long_options[] = {
        {"date", required_argument, 0, 'd'},
        {"lat", required_argument, 0, 'a'},
        {"lon", required_argument, 0, 'o'},
        {"phase-only", no_argument, 0, 'p'},
        {"save-location", required_argument, 0, 's'},
        {"use-location", required_argument, 0, 'u'},
        {"list-locations", no_argument, 0, 'l'},
        {0,0,0,0}
    };
    int opt;
    string dateStr, saveLocation, useLocation;
    double lat = DEFAULT_LAT, lon = DEFAULT_LON;
    bool phaseOnly = false, listLocations = false;

    while ((opt = getopt_long(argc, argv, "d:a:o:ps:u:l", long_options, nullptr)) != -1) {
        switch (opt) {
            case 'd': dateStr = optarg; break;
            case 'a': lat = stod(optarg); break;
            case 'o': lon = stod(optarg); break;
            case 'p': phaseOnly = true; break;
            case 's': saveLocation = optarg; break;
            case 'u': useLocation = optarg; break;
            case 'l': listLocations = true; break;
            default:
                cerr << "Usage: lunar_fishing --date YYYY-MM-DD --lat LAT --lon LON --phase-only --save-location NAME --use-location NAME --list-locations\n";
                return 1;
        }
    }

    // Load config
    json config;
    ifstream f(CONFIG_FILE);
    if (f.is_open()) {
        f >> config;
    }

    if (listLocations) {
        if (!config.contains("locations") || config["locations"].empty()) {
            cout << "No saved locations.\n";
            return 0;
        }
        cout << "\n📍 Saved Locations:\n";
        for (auto& [name, loc] : config["locations"].items()) {
            string latStr = to_string(abs(loc["lat"].get<double>())).substr(0,5) + "°" + (loc["lat"].get<double>() >= 0 ? "N" : "S");
            string lonStr = to_string(abs(loc["lon"].get<double>())).substr(0,5) + "°" + (loc["lon"].get<double>() >= 0 ? "E" : "W");
            cout << "  " << name << ": " << latStr << ", " << lonStr << "\n";
        }
        return 0;
    }

    if (!saveLocation.empty()) {
        config["locations"][saveLocation] = {{"lat", lat}, {"lon", lon}};
        ofstream out(CONFIG_FILE);
        out << setw(2) << config << endl;
        cout << "✅ Location '" << saveLocation << "' saved.\n";
    }

    if (!useLocation.empty()) {
        if (config.contains("locations") && config["locations"].contains(useLocation)) {
            lat = config["locations"][useLocation]["lat"];
            lon = config["locations"][useLocation]["lon"];
            cout << "📍 Using saved location: " << useLocation << "\n";
        } else {
            cout << "Location '" << useLocation << "' not found.\n";
            return 1;
        }
    }

    time_t now = time(nullptr);
    tm dt = *gmtime(&now);
    if (!dateStr.empty()) {
        strptime(dateStr.c_str(), "%Y-%m-%d", &dt);
    }

    render(dt, lat, lon, phaseOnly);
    return 0;
}
