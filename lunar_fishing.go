// lunar_fishing.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"time"
)

const (
	DEFAULT_LAT = 0.0
	DEFAULT_LON = 0.0
	CONFIG_FILE = "lunar_fishing_config.json"
)

type Config struct {
	Locations map[string]struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"locations"`
}

type MoonData struct {
	Phase        string
	Illumination float64
	Age          float64
	Zodiac       string
}

type SolunarData struct {
	Major1Start  float64
	Major1End    float64
	Major2Start  float64
	Major2End    float64
	Minor1Start  float64
	Minor1End    float64
	Minor2Start  float64
	Minor2End    float64
	MoonTransit  float64
}

func julianDay(t time.Time) float64 {
	year := t.Year()
	month := int(t.Month())
	day := float64(t.Day()) + float64(t.Hour())/24.0 + float64(t.Minute())/1440.0 + float64(t.Second())/86400.0
	if month <= 2 {
		year--
		month += 12
	}
	A := year / 100
	B := 2 - A + A/4
	return float64(int(365.25*float64(year+4716))) + float64(int(30.6001*float64(month+1))) + day + float64(B) - 1524.5
}

func moonPosition(jd float64) (float64, float64) {
	T := (jd - 2451545.0) / 36525.0
	L_prime := 218.3165 + 481267.8813*T
	D := 297.8502 + 445267.1114*T
	M := 357.5291 + 35999.0503*T
	M_prime := 134.9634 + 477198.8676*T
	F := 93.2720 + 483202.0175*T

	L_prime = math.Mod(L_prime, 360) * math.Pi / 180
	D = math.Mod(D, 360) * math.Pi / 180
	M = math.Mod(M, 360) * math.Pi / 180
	M_prime = math.Mod(M_prime, 360) * math.Pi / 180
	F = math.Mod(F, 360) * math.Pi / 180

	lon := L_prime + (6.289*math.Sin(M_prime)+1.274*math.Sin(2*D-M_prime)+0.658*math.Sin(2*D)+0.214*math.Sin(2*M_prime)-0.186*math.Sin(M)-0.114*math.Sin(2*F))*math.Pi/180
	lat := (5.128*math.Sin(F)+0.280*math.Sin(M_prime+F)+0.278*math.Sin(M_prime-F)+0.173*math.Sin(2*D-F))*math.Pi/180
	return lon, lat
}

func sunPosition(jd float64) float64 {
	T := (jd - 2451545.0) / 36525.0
	M := math.Mod(357.5291+35999.0503*T, 360) * math.Pi / 180
	C := 1.9146*math.Sin(M) + 0.0200*math.Sin(2*M) + 0.0003*math.Sin(3*M)
	lon := math.Mod(280.4665+36000.7698*T+C, 360) * math.Pi / 180
	return lon
}

func greenwichSiderealTime(jd float64) float64 {
	T := (jd - 2451545.0) / 36525.0
	gmst := 280.46061837 + 360.98564736629*(jd-2451545.0) + 0.000387933*T*T - T*T*T/38710000.0
	gmst = math.Mod(gmst, 360.0)
	return gmst / 15.0
}

func moonPhase(t time.Time) MoonData {
	jd := julianDay(t)
	lonMoon, _ := moonPosition(jd)
	lonSun := sunPosition(jd)

	elong := lonMoon - lonSun
	elong = math.Atan2(math.Sin(elong), math.Cos(elong))
	phaseAngle := math.Atan2(math.Sin(elong), math.Cos(elong))
	illumination := (1 + math.Cos(phaseAngle)) / 2

	age := (jd - 2451550.1) / 29.53058867
	age = math.Mod(age, 29.53058867)
	if age < 0 {
		age += 29.53058867
	}

	var phase string
	if age < 1.0 {
		phase = "New Moon"
	} else if age < 7.38 {
		phase = "Waxing Crescent"
	} else if age < 8.38 {
		phase = "First Quarter"
	} else if age < 14.77 {
		phase = "Waxing Gibbous"
	} else if age < 15.77 {
		phase = "Full Moon"
	} else if age < 22.15 {
		phase = "Waning Gibbous"
	} else if age < 23.15 {
		phase = "Last Quarter"
	} else {
		phase = "Waning Crescent"
	}

	signs := []string{"Aries", "Taurus", "Gemini", "Cancer", "Leo", "Virgo", "Libra", "Scorpio", "Sagittarius", "Capricorn", "Aquarius", "Pisces"}
	lonDeg := math.Mod(lonMoon*180/math.Pi, 360)
	if lonDeg < 0 {
		lonDeg += 360
	}
	idx := int(lonDeg / 30)
	zodiac := signs[idx]

	return MoonData{Phase: phase, Illumination: illumination * 100, Age: age, Zodiac: zodiac}
}

func solunar(t time.Time, lat, lon float64) SolunarData {
	jd := julianDay(t)
	lonMoon, _ := moonPosition(jd)
	lonMoonDeg := math.Mod(lonMoon*180/math.Pi, 360)
	if lonMoonDeg < 0 {
		lonMoonDeg += 360
	}
	gst := greenwichSiderealTime(jd)
	lst := math.Mod(gst+lon/15.0, 24.0)
	moonRA := lonMoonDeg / 15.0
	moonTransit := math.Mod(moonRA-lst, 24.0)
	if moonTransit < 0 {
		moonTransit += 24.0
	}
	// Major: transit ± 1h, opposite ± 1h
	major1Start := moonTransit - 1.0
	major1End := moonTransit + 1.0
	major2Start := moonTransit + 12.0 - 1.0
	major2End := moonTransit + 12.0 + 1.0
	// Minor: ± 30 min around 6h and 18h offsets
	minor1Start := moonTransit + 6.0 - 0.5
	minor1End := moonTransit + 6.0 + 0.5
	minor2Start := moonTransit + 18.0 - 0.5
	minor2End := moonTransit + 18.0 + 0.5

	return SolunarData{
		Major1Start: major1Start, Major1End: major1End,
		Major2Start: major2Start, Major2End: major2End,
		Minor1Start: minor1Start, Minor1End: minor1End,
		Minor2Start: minor2Start, Minor2End: minor2End,
		MoonTransit: moonTransit,
	}
}

func fishingRating(phase string) int {
	switch phase {
	case "Full Moon", "New Moon":
		return 4
	case "Waxing Gibbous", "Waning Gibbous", "First Quarter", "Last Quarter":
		return 3
	case "Waxing Crescent", "Waning Crescent":
		return 2
	default:
		return 1
	}
}

func render(t time.Time, lat, lon float64, phaseOnly bool) {
	moonData := moonPhase(t)
	solunarData := solunar(t, lat, lon)
	rating := fishingRating(moonData.Phase)

	if phaseOnly {
		fmt.Println(moonData.Phase)
		return
	}

	ratings := []string{"Poor", "Fair", "Good", "Excellent"}
	ratingEmoji := []string{"⭐", "⭐⭐", "⭐⭐⭐", "⭐⭐⭐⭐"}

	fmt.Printf("\n🎣 Lunar Fishing Calendar\n")
	latStr := fmt.Sprintf("%.2f°%c", math.Abs(lat), 'N')
	if lat < 0 {
		latStr = fmt.Sprintf("%.2f°%c", math.Abs(lat), 'S')
	}
	lonStr := fmt.Sprintf("%.2f°%c", math.Abs(lon), 'E')
	if lon < 0 {
		lonStr = fmt.Sprintf("%.2f°%c", math.Abs(lon), 'W')
	}
	fmt.Printf("Location: %s, %s\n", latStr, lonStr)
	fmt.Printf("Date: %s\n", t.Format("2006-01-02 15:04"))

	fmt.Printf("\n🌓 Moon Phase: %s (%.1f%% illuminated)\n", moonData.Phase, moonData.Illumination)
	fmt.Printf("🌙 Moon Age: %.1f days\n", moonData.Age)
	fmt.Printf("♑ Zodiac: %s\n", moonData.Zodiac)

	fmt.Printf("\n🎯 Solunar Feeding Periods:\n")
	// Major 1
	if solunarData.Major1Start >= 0 && solunarData.Major1Start < 24 {
		fmt.Printf("  Major Period 1: %02d:%02d – %02d:%02d (%s %s)\n",
			int(solunarData.Major1Start), int((solunarData.Major1Start*60)%60),
			int(solunarData.Major1End), int((solunarData.Major1End*60)%60),
			ratingEmoji[rating-1], ratings[rating-1])
	}
	if solunarData.Major2Start >= 0 && solunarData.Major2Start < 24 {
		fmt.Printf("  Major Period 2: %02d:%02d – %02d:%02d (%s %s)\n",
			int(solunarData.Major2Start), int((solunarData.Major2Start*60)%60),
			int(solunarData.Major2End), int((solunarData.Major2End*60)%60),
			ratingEmoji[rating-1], ratings[rating-1])
	}
	if solunarData.Minor1Start >= 0 && solunarData.Minor1Start < 24 {
		fmt.Printf("  Minor Period 1: %02d:%02d – %02d:%02d (⭐ Good)\n",
			int(solunarData.Minor1Start), int((solunarData.Minor1Start*60)%60),
			int(solunarData.Minor1End), int((solunarData.Minor1End*60)%60))
	}
	if solunarData.Minor2Start >= 0 && solunarData.Minor2Start < 24 {
		fmt.Printf("  Minor Period 2: %02d:%02d – %02d:%02d (⭐ Good)\n",
			int(solunarData.Minor2Start), int((solunarData.Minor2Start*60)%60),
			int(solunarData.Minor2End), int((solunarData.Minor2End*60)%60))
	}

	fmt.Printf("\n⭐ Best Fishing Rating: %s (%s)\n", ratings[rating-1], ratingEmoji[rating-1])

	// Sunrise/sunset (simplified)
	dayOfYear := t.YearDay()
	declination := 23.44 * math.Sin((284+dayOfYear)*360*math.Pi/180/365)
	latRad := lat * math.Pi / 180
	decRad := declination * math.Pi / 180
	cosHA := -math.Tan(latRad) * math.Tan(decRad)
	var ha float64
	if cosHA < -1 {
		ha = math.Pi
	} else if cosHA > 1 {
		ha = 0
	} else {
		ha = math.Acos(cosHA)
	}
	dayLength := ha * 2 / (math.Pi / 12)
	noon := 12.0 - lon/15.0
	sunrise := noon - dayLength/2
	sunset := noon + dayLength/2
	sunriseH := int(sunrise)
	sunriseM := int((sunrise - float64(sunriseH)) * 60)
	sunsetH := int(sunset)
	sunsetM := int((sunset - float64(sunsetH)) * 60)
	fmt.Printf("🌅 Sunrise: %02d:%02d | 🌇 Sunset: %02d:%02d\n", sunriseH, sunriseM, sunsetH, sunsetM)

	moonRise := solunarData.MoonTransit + 6
	moonSet := solunarData.MoonTransit + 18
	if moonRise >= 24 {
		moonRise -= 24
	}
	if moonSet >= 24 {
		moonSet -= 24
	}
	moonRiseH := int(moonRise)
	moonRiseM := int((moonRise - float64(moonRiseH)) * 60)
	moonSetH := int(moonSet)
	moonSetM := int((moonSet - float64(moonSetH)) * 60)
	fmt.Printf("🌙 Moonrise: %02d:%02d | 🌇 Moonset: %02d:%02d\n", moonRiseH, moonRiseM, moonSetH, moonSetM)
}

func main() {
	var (
		dateStr      = flag.String("date", "", "YYYY-MM-DD")
		lat          = flag.Float64("lat", DEFAULT_LAT, "Latitude (positive North)")
		lon          = flag.Float64("lon", DEFAULT_LON, "Longitude (positive East)")
		phaseOnly    = flag.Bool("phase-only", false, "Output only phase name")
		saveLocation = flag.String("save-location", "", "Save location with name")
		useLocation  = flag.String("use-location", "", "Use saved location")
		listLocations = flag.Bool("list-locations", false, "List saved locations")
	)
	flag.Parse()

	var config Config
	if data, err := os.ReadFile(CONFIG_FILE); err == nil {
		json.Unmarshal(data, &config)
	}
	if config.Locations == nil {
		config.Locations = make(map[string]struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		})
	}

	if *listLocations {
		if len(config.Locations) == 0 {
			fmt.Println("No saved locations.")
			return
		}
		fmt.Println("\n📍 Saved Locations:")
		for name, loc := range config.Locations {
			latStr := fmt.Sprintf("%.2f°%c", math.Abs(loc.Lat), 'N')
			if loc.Lat < 0 {
				latStr = fmt.Sprintf("%.2f°%c", math.Abs(loc.Lat), 'S')
			}
			lonStr := fmt.Sprintf("%.2f°%c", math.Abs(loc.Lon), 'E')
			if loc.Lon < 0 {
				lonStr = fmt.Sprintf("%.2f°%c", math.Abs(loc.Lon), 'W')
			}
			fmt.Printf("  %s: %s, %s\n", name, latStr, lonStr)
		}
		return
	}

	if *saveLocation != "" {
		config.Locations[*saveLocation] = struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		}{*lat, *lon}
		data, _ := json.MarshalIndent(config, "", "  ")
		os.WriteFile(CONFIG_FILE, data, 0644)
		fmt.Printf("✅ Location '%s' saved.\n", *saveLocation)
	}

	useLat := *lat
	useLon := *lon
	if *useLocation != "" {
		if loc, ok := config.Locations[*useLocation]; ok {
			useLat = loc.Lat
			useLon = loc.Lon
			fmt.Printf("📍 Using saved location: %s\n", *useLocation)
		} else {
			fmt.Printf("Location '%s' not found.\n", *useLocation)
			return
		}
	}

	t := time.Now()
	if *dateStr != "" {
		t, _ = time.Parse("2006-01-02", *dateStr)
	}
	render(t, useLat, useLon, *phaseOnly)
}
