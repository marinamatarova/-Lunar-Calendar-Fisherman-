🎣 Lunar Calendar (Fisherman) — Multi‑Language Solunar Fishing Guide
8 languages, one comprehensive fishing planner – track moon phases, predict fish activity, and find the best fishing days using Solunar theory – right from your terminal.

✨ Features
🌓 Moon phase – show phase, illumination, age, and zodiac sign

🎣 Solunar predictions – major and minor feeding periods (peak fish activity)

⭐ Best fishing days – rating system based on moon phase and solunar peaks

🌅 Sunrise/Sunset – approximate times for your location

🌙 Moonrise/Moonset – for perfect timing

📍 Location support – use latitude/longitude for accurate times

💾 Save favorite spots – persistent storage of fishing locations

📊 Daily rating – Poor, Fair, Good, Excellent

🚀 Quick Start
All implementations follow the same CLI pattern:

bash
# Show today's fishing forecast for your saved location
<command>

# Show forecast for a specific date
<command> --date 2026-08-25

# Use a specific location
<command> --lat 51.5074 --lon -0.1276

# Save a location for future use
<command> --save-location "Thames River" --lat 51.5074 --lon -0.1276

# Show the current moon phase only
<command> --phase-only

# List all saved locations
<command> --list-locations
Arguments:

--date YYYY-MM-DD – date to forecast (default: today)

--lat <degrees> – latitude (positive North)

--lon <degrees> – longitude (positive East)

--phase-only – output only the phase name

--save-location <name> – save location with a name

--list-locations – show all saved spots

--use-location <name> – use a saved location

📸 Example Output
text
🎣 Lunar Fishing Calendar
Location: Thames River (51.51°N, 0.13°W)
Date: 2026-08-25 06:30

🌓 Moon Phase: Waxing Gibbous (72.3% illuminated)
🌙 Moon Age: 9.7 days
♑ Zodiac: Sagittarius

🎯 Solunar Feeding Periods:
  Major Period 1: 04:30 – 06:30 (⭐⭐⭐ Excellent)
  Minor Period 1: 10:45 – 11:45 (⭐ Good)
  Major Period 2: 16:50 – 18:50 (⭐⭐⭐ Excellent)
  Minor Period 2: 23:10 – 00:10 (⭐ Good)

⭐ Best Fishing Rating: Excellent
🌅 Sunrise: 06:03 | 🌇 Sunset: 20:15
🌙 Moonrise: 14:22 | 🌇 Moonset: 02:45

📋 Daily Tips:
  • Best times: 04:30–06:30 and 16:50–18:50
  • Moon phase favors active feeding
  • Try topwater lures during peak periods
📁 Repository Structure
text
.
├── README.md
├── python/
│   └── lunar_fishing.py
├── go/
│   └── lunar_fishing.go
├── javascript/
│   └── lunar_fishing.js
├── ruby/
│   └── lunar_fishing.rb
├── php/
│   └── lunar_fishing.php
├── java/
│   └── LunarFishing.java
├── csharp/
│   └── LunarFishing.cs
└── cpp/
    └── lunar_fishing.cpp
