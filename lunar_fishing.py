# lunar_fishing.py
import math
import json
import os
import argparse
from datetime import datetime, timedelta

CONFIG_FILE = "lunar_fishing_config.json"
DEFAULT_LAT = 0.0
DEFAULT_LON = 0.0

class LunarFishing:
    def __init__(self, lat=DEFAULT_LAT, lon=DEFAULT_LON):
        self.lat = lat
        self.lon = lon
        self.config = self.load_config()

    def load_config(self):
        if os.path.exists(CONFIG_FILE):
            with open(CONFIG_FILE, "r") as f:
                return json.load(f)
        return {"locations": {}}

    def save_config(self):
        with open(CONFIG_FILE, "w") as f:
            json.dump(self.config, f, indent=2)

    def save_location(self, name, lat, lon):
        self.config["locations"][name] = {"lat": lat, "lon": lon}
        self.save_config()

    def julian_day(self, dt):
        year = dt.year
        month = dt.month
        day = dt.day + dt.hour/24.0 + dt.minute/1440.0 + dt.second/86400.0
        if month <= 2:
            year -= 1
            month += 12
        A = int(year / 100)
        B = 2 - A + int(A / 4)
        return int(365.25 * (year + 4716)) + int(30.6001 * (month + 1)) + day + B - 1524.5

    def moon_position(self, jd):
        T = (jd - 2451545.0) / 36525.0
        L_prime = 218.3165 + 481267.8813 * T
        D = 297.8502 + 445267.1114 * T
        M = 357.5291 + 35999.0503 * T
        M_prime = 134.9634 + 477198.8676 * T
        F = 93.2720 + 483202.0175 * T

        L_prime = ((L_prime % 360) + 360) % 360 * math.pi / 180
        D = ((D % 360) + 360) % 360 * math.pi / 180
        M = ((M % 360) + 360) % 360 * math.pi / 180
        M_prime = ((M_prime % 360) + 360) % 360 * math.pi / 180
        F = ((F % 360) + 360) % 360 * math.pi / 180

        lon = L_prime + (6.289 * math.sin(M_prime) + 1.274 * math.sin(2*D - M_prime) +
                         0.658 * math.sin(2*D) + 0.214 * math.sin(2*M_prime) -
                         0.186 * math.sin(M) - 0.114 * math.sin(2*F)) * math.pi / 180
        lat = (5.128 * math.sin(F) + 0.280 * math.sin(M_prime + F) +
               0.278 * math.sin(M_prime - F) + 0.173 * math.sin(2*D - F)) * math.pi / 180
        return lon, lat

    def sun_position(self, jd):
        T = (jd - 2451545.0) / 36525.0
        M = ((357.5291 + 35999.0503 * T) % 360) * math.pi / 180
        C = 1.9146 * math.sin(M) + 0.0200 * math.sin(2*M) + 0.0003 * math.sin(3*M)
        lon = ((280.4665 + 36000.7698 * T + C) % 360) * math.pi / 180
        return lon

    def moon_phase(self, dt):
        jd = self.julian_day(dt)
        lon_moon, _ = self.moon_position(jd)
        lon_sun = self.sun_position(jd)

        elong = lon_moon - lon_sun
        elong = math.atan2(math.sin(elong), math.cos(elong))
        phase_angle = math.atan2(math.sin(elong), math.cos(elong))
        illumination = (1 + math.cos(phase_angle)) / 2

        age = (jd - 2451550.1) / 29.53058867
        age = ((age % 29.53058867) + 29.53058867) % 29.53058867

        if age < 1.0:
            phase = "New Moon"
        elif age < 7.38:
            phase = "Waxing Crescent"
        elif age < 8.38:
            phase = "First Quarter"
        elif age < 14.77:
            phase = "Waxing Gibbous"
        elif age < 15.77:
            phase = "Full Moon"
        elif age < 22.15:
            phase = "Waning Gibbous"
        elif age < 23.15:
            phase = "Last Quarter"
        else:
            phase = "Waning Crescent"

        signs = ["Aries", "Taurus", "Gemini", "Cancer", "Leo", "Virgo",
                 "Libra", "Scorpio", "Sagittarius", "Capricorn", "Aquarius", "Pisces"]
        lon_deg = ((lon_moon * 180 / math.pi) % 360 + 360) % 360
        idx = int(lon_deg / 30)
        zodiac = signs[idx]

        return {"phase": phase, "illumination": illumination * 100,
                "age": age, "zodiac": zodiac, "jd": jd, "lon_moon": lon_moon}

    def solunar(self, dt, lat, lon):
        """Calculate solunar feeding periods."""
        # Simplified solunar: major periods occur when moon is overhead/underfoot
        # Minor periods occur when moon is rising/setting
        # For a real implementation, we'd calculate exact lunar transit times
        jd = self.julian_day(dt)
        lon_moon, _ = self.moon_position(jd)
        # Convert to degrees
        lon_moon_deg = lon_moon * 180 / math.pi % 360
        # Moon transit time (approx)
        # Local Sidereal Time
        gst = self.greenwich_sidereal_time(jd)
        lst = (gst + lon / 15.0) % 24.0
        # Moon RA (approx from lon_moon)
        moon_ra = lon_moon_deg / 15.0
        moon_transit = (moon_ra - lst) * 24.0 / 360.0
        # Major periods: ± 1 hour around transit and opposite
        major1_start = dt.replace(hour=0, minute=0, second=0) + timedelta(hours=moon_transit - 1)
        major1_end = dt.replace(hour=0, minute=0, second=0) + timedelta(hours=moon_transit + 1)
        major2_start = dt.replace(hour=0, minute=0, second=0) + timedelta(hours=moon_transit + 12 - 1)
        major2_end = dt.replace(hour=0, minute=0, second=0) + timedelta(hours=moon_transit + 12 + 1)
        # Minor periods: ± 30 min around moonrise/set (simplified)
        # For simplicity, use 6h offset from transit
        minor1_start = dt.replace(hour=0, minute=0, second=0) + timedelta(hours=moon_transit + 6 - 0.5)
        minor1_end = dt.replace(hour=0, minute=0, second=0) + timedelta(hours=moon_transit + 6 + 0.5)
        minor2_start = dt.replace(hour=0, minute=0, second=0) + timedelta(hours=moon_transit + 18 - 0.5)
        minor2_end = dt.replace(hour=0, minute=0, second=0) + timedelta(hours=moon_transit + 18 + 0.5)
        return {
            "major1": (major1_start, major1_end),
            "major2": (major2_start, major2_end),
            "minor1": (minor1_start, minor1_end),
            "minor2": (minor2_start, minor2_end),
            "moon_transit": moon_transit
        }

    def greenwich_sidereal_time(self, jd):
        T = (jd - 2451545.0) / 36525.0
        gmst = 280.46061837 + 360.98564736629 * (jd - 2451545.0) + 0.000387933 * T*T - T*T*T / 38710000.0
        gmst = gmst % 360.0
        return gmst / 15.0

    def sun_rise_set(self, dt, lat, lon):
        """Simplified sunrise/sunset times."""
        # Placeholder: real implementation would use more complex calculations
        # For now, return approximate times based on latitude and date
        # This is a very rough approximation
        day_of_year = dt.timetuple().tm_yday
        declination = 23.44 * math.sin((284 + day_of_year) * 360 * math.pi / 180 / 365)
        lat_rad = lat * math.pi / 180
        dec_rad = declination * math.pi / 180
        cos_ha = -math.tan(lat_rad) * math.tan(dec_rad)
        if cos_ha < -1:
            ha = math.pi
        elif cos_ha > 1:
            ha = 0
        else:
            ha = math.acos(cos_ha)
        day_length = ha * 2 / (math.pi / 12)
        noon = 12.0 - lon / 15.0
        sunrise = noon - day_length / 2
        sunset = noon + day_length / 2
        return sunrise, sunset

    def fishing_rating(self, moon_phase_data, solunar_data):
        """Calculate fishing rating based on moon phase and solunar peaks."""
        phase = moon_phase_data["phase"]
        illumination = moon_phase_data["illumination"]
        # Rating: Excellent if Full Moon or New Moon with good solunar timing
        # Good: Waxing/Waning Gibbous, First/Last Quarter with peaks
        # Fair: Crescent phases
        # Poor: no peaks
        if phase in ["Full Moon", "New Moon"]:
            base = 4  # Excellent
        elif phase in ["Waxing Gibbous", "Waning Gibbous", "First Quarter", "Last Quarter"]:
            base = 3  # Good
        elif phase in ["Waxing Crescent", "Waning Crescent"]:
            base = 2  # Fair
        else:
            base = 1  # Poor
        # Check if any major peaks fall within reasonable hours (6-18)
        # Add bonus for major periods in daytime
        return base

    def format_time(self, dt):
        return dt.strftime("%H:%M")

    def render(self, dt, lat, lon, phase_only=False):
        moon_data = self.moon_phase(dt)
        solunar = self.solunar(dt, lat, lon)
        rating = self.fishing_rating(moon_data, solunar)
        sunrise, sunset = self.sun_rise_set(dt, lat, lon)

        if phase_only:
            print(moon_data["phase"])
            return

        print(f"\n🎣 Lunar Fishing Calendar")
        lat_str = f"{abs(lat):.2f}°{'N' if lat >=0 else 'S'}"
        lon_str = f"{abs(lon):.2f}°{'E' if lon >=0 else 'W'}"
        print(f"Location: {lat_str}, {lon_str}")
        print(f"Date: {dt.strftime('%Y-%m-%d %H:%M')}")

        print(f"\n🌓 Moon Phase: {moon_data['phase']} ({moon_data['illumination']:.1f}% illuminated)")
        print(f"🌙 Moon Age: {moon_data['age']:.1f} days")
        print(f"♑ Zodiac: {moon_data['zodiac']}")

        print(f"\n🎯 Solunar Feeding Periods:")
        ratings = ["Poor", "Fair", "Good", "Excellent"]
        rating_emoji = ["⭐", "⭐⭐", "⭐⭐⭐", "⭐⭐⭐⭐"]
        # Major periods
        for i, (start, end) in enumerate([(solunar["major1"], solunar["major2"])], 1):
            if start.hour < 24:
                print(f"  Major Period {i}: {self.format_time(start)} – {self.format_time(end)} ({rating_emoji[rating-1]} {ratings[rating-1]})")
        for i, (start, end) in enumerate([(solunar["minor1"], solunar["minor2"])], 1):
            if start.hour < 24:
                print(f"  Minor Period {i}: {self.format_time(start)} – {self.format_time(end)} (⭐ Good)")

        print(f"\n⭐ Best Fishing Rating: {ratings[rating-1]} ({rating_emoji[rating-1]})")
        sunrise_h = int(sunrise)
        sunrise_m = int((sunrise - sunrise_h) * 60)
        sunset_h = int(sunset)
        sunset_m = int((sunset - sunset_h) * 60)
        print(f"🌅 Sunrise: {sunrise_h:02d}:{sunrise_m:02d} | 🌇 Sunset: {sunset_h:02d}:{sunset_m:02d}")

        # Moon rise/set (approx)
        moon_rise_h = (solunar["moon_transit"] + 6) % 24
        moon_set_h = (solunar["moon_transit"] + 18) % 24
        print(f"🌙 Moonrise: {int(moon_rise_h):02d}:{int((moon_rise_h%1)*60):02d} | 🌇 Moonset: {int(moon_set_h):02d}:{int((moon_set_h%1)*60):02d}")

def main():
    parser = argparse.ArgumentParser(description="Lunar Fishing Calendar")
    parser.add_argument("--date", help="YYYY-MM-DD")
    parser.add_argument("--lat", type=float, help="Latitude (positive North)")
    parser.add_argument("--lon", type=float, help="Longitude (positive East)")
    parser.add_argument("--phase-only", action="store_true", help="Output only phase name")
    parser.add_argument("--save-location", help="Save location with name")
    parser.add_argument("--use-location", help="Use saved location")
    parser.add_argument("--list-locations", action="store_true", help="List saved locations")
    args = parser.parse_args()

    app = LunarFishing()

    if args.list_locations:
        if not app.config["locations"]:
            print("No saved locations.")
            return
        print("\n📍 Saved Locations:")
        for name, data in app.config["locations"].items():
            lat = data["lat"]
            lon = data["lon"]
            print(f"  {name}: {lat:.2f}°{'N' if lat>=0 else 'S'}, {lon:.2f}°{'E' if lon>=0 else 'W'}")
        return

    if args.save_location and args.lat is not None and args.lon is not None:
        app.save_location(args.save_location, args.lat, args.lon)
        print(f"✅ Location '{args.save_location}' saved.")

    lat = args.lat if args.lat is not None else DEFAULT_LAT
    lon = args.lon if args.lon is not None else DEFAULT_LON

    if args.use_location and args.use_location in app.config["locations"]:
        loc = app.config["locations"][args.use_location]
        lat = loc["lat"]
        lon = loc["lon"]
        print(f"📍 Using saved location: {args.use_location}")

    dt = datetime.now()
    if args.date:
        dt = datetime.strptime(args.date, "%Y-%m-%d")

    app.render(dt, lat, lon, args.phase_only)

if __name__ == "__main__":
    main()
