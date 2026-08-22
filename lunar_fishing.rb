# lunar_fishing.rb
#!/usr/bin/env ruby
require 'json'
require 'date'
require 'optparse'

DEFAULT_LAT = 0.0
DEFAULT_LON = 0.0
CONFIG_FILE = 'lunar_fishing_config.json'

def julian_day(dt)
  year = dt.year
  month = dt.month
  day = dt.day + dt.hour/24.0 + dt.min/1440.0 + dt.sec/86400.0
  if month <= 2
    year -= 1
    month += 12
  end
  a = (year / 100).to_i
  b = 2 - a + (a / 4).to_i
  (365.25 * (year + 4716)).to_i + (30.6001 * (month + 1)).to_i + day + b - 1524.5
end

def moon_position(jd)
  t = (jd - 2451545.0) / 36525.0
  l_prime = 218.3165 + 481267.8813 * t
  d = 297.8502 + 445267.1114 * t
  m = 357.5291 + 35999.0503 * t
  m_prime = 134.9634 + 477198.8676 * t
  f = 93.2720 + 483202.0175 * t

  l_prime = (l_prime % 360) * Math::PI / 180
  d = (d % 360) * Math::PI / 180
  m = (m % 360) * Math::PI / 180
  m_prime = (m_prime % 360) * Math::PI / 180
  f = (f % 360) * Math::PI / 180

  lon = l_prime + (6.289 * Math.sin(m_prime) + 1.274 * Math.sin(2*d - m_prime) + 0.658 * Math.sin(2*d) + 0.214 * Math.sin(2*m_prime) - 0.186 * Math.sin(m) - 0.114 * Math.sin(2*f)) * Math::PI / 180
  lat = (5.128 * Math.sin(f) + 0.280 * Math.sin(m_prime + f) + 0.278 * Math.sin(m_prime - f) + 0.173 * Math.sin(2*d - f)) * Math::PI / 180
  [lon, lat]
end

def sun_position(jd)
  t = (jd - 2451545.0) / 36525.0
  m = (357.5291 + 35999.0503 * t) % 360 * Math::PI / 180
  c = 1.9146 * Math.sin(m) + 0.0200 * Math.sin(2*m) + 0.0003 * Math.sin(3*m)
  lon = (280.4665 + 36000.7698 * t + c) % 360 * Math::PI / 180
  lon
end

def greenwich_sidereal_time(jd)
  t = (jd - 2451545.0) / 36525.0
  gmst = 280.46061837 + 360.98564736629 * (jd - 2451545.0) + 0.000387933 * t*t - t*t*t / 38710000.0
  gmst = gmst % 360.0
  gmst / 15.0
end

def moon_phase(dt)
  jd = julian_day(dt)
  lon_moon, _ = moon_position(jd)
  lon_sun = sun_position(jd)

  elong = lon_moon - lon_sun
  elong = Math.atan2(Math.sin(elong), Math.cos(elong))
  phase_angle = Math.atan2(Math.sin(elong), Math.cos(elong))
  illumination = (1 + Math.cos(phase_angle)) / 2

  age = (jd - 2451550.1) / 29.53058867
  age = age % 29.53058867

  phase = case age
          when 0...1.0 then "New Moon"
          when 1.0...7.38 then "Waxing Crescent"
          when 7.38...8.38 then "First Quarter"
          when 8.38...14.77 then "Waxing Gibbous"
          when 14.77...15.77 then "Full Moon"
          when 15.77...22.15 then "Waning Gibbous"
          when 22.15...23.15 then "Last Quarter"
          else "Waning Crescent"
          end

  signs = ["Aries","Taurus","Gemini","Cancer","Leo","Virgo","Libra","Scorpio","Sagittarius","Capricorn","Aquarius","Pisces"]
  lon_deg = (lon_moon * 180 / Math::PI) % 360
  idx = (lon_deg / 30).to_i
  zodiac = signs[idx]

  { phase: phase, illumination: illumination * 100, age: age, zodiac: zodiac, lon_moon: lon_moon, jd: jd }
end

def solunar(dt, lat, lon)
  jd = julian_day(dt)
  lon_moon, _ = moon_position(jd)
  lon_moon_deg = (lon_moon * 180 / Math::PI) % 360
  gst = greenwich_sidereal_time(jd)
  lst = (gst + lon / 15.0) % 24.0
  moon_ra = lon_moon_deg / 15.0
  moon_transit = (moon_ra - lst) % 24.0
  {
    major1_start: moon_transit - 1.0,
    major1_end: moon_transit + 1.0,
    major2_start: moon_transit + 12.0 - 1.0,
    major2_end: moon_transit + 12.0 + 1.0,
    minor1_start: moon_transit + 6.0 - 0.5,
    minor1_end: moon_transit + 6.0 + 0.5,
    minor2_start: moon_transit + 18.0 - 0.5,
    minor2_end: moon_transit + 18.0 + 0.5,
    moon_transit: moon_transit
  }
end

def fishing_rating(phase)
  case phase
  when "Full Moon", "New Moon" then 4
  when "Waxing Gibbous", "Waning Gibbous", "First Quarter", "Last Quarter" then 3
  when "Waxing Crescent", "Waning Crescent" then 2
  else 1
  end
end

def format_time(hours)
  h = hours.to_i
  m = ((hours - h) * 60).to_i
  sprintf("%02d:%02d", h, m)
end

def render(dt, lat, lon, phase_only)
  moon_data = moon_phase(dt)
  solunar_data = solunar(dt, lat, lon)
  rating = fishing_rating(moon_data[:phase])

  if phase_only
    puts moon_data[:phase]
    return
  end

  ratings = ["Poor", "Fair", "Good", "Excellent"]
  rating_emoji = ["⭐", "⭐⭐", "⭐⭐⭐", "⭐⭐⭐⭐"]

  puts "\n🎣 Lunar Fishing Calendar"
  lat_str = "#{lat.abs.round(2)}°#{lat >= 0 ? 'N' : 'S'}"
  lon_str = "#{lon.abs.round(2)}°#{lon >= 0 ? 'E' : 'W'}"
  puts "Location: #{lat_str}, #{lon_str}"
  puts "Date: #{dt.strftime('%Y-%m-%d %H:%M')}"

  puts "\n🌓 Moon Phase: #{moon_data[:phase]} (#{moon_data[:illumination].round(1)}% illuminated)"
  puts "🌙 Moon Age: #{moon_data[:age].round(1)} days"
  puts "♑ Zodiac: #{moon_data[:zodiac]}"

  puts "\n🎯 Solunar Feeding Periods:"
  periods = [
    { start: solunar_data[:major1_start], end: solunar_data[:major1_end], label: "Major Period 1", emoji: rating_emoji[rating-1], rating: ratings[rating-1] },
    { start: solunar_data[:major2_start], end: solunar_data[:major2_end], label: "Major Period 2", emoji: rating_emoji[rating-1], rating: ratings[rating-1] },
    { start: solunar_data[:minor1_start], end: solunar_data[:minor1_end], label: "Minor Period 1", emoji: "⭐", rating: "Good" },
    { start: solunar_data[:minor2_start], end: solunar_data[:minor2_end], label: "Minor Period 2", emoji: "⭐", rating: "Good" },
  ]
  periods.each do |p|
    if p[:start] >= 0 && p[:start] < 24
      puts "  #{p[:label]}: #{format_time(p[:start])} – #{format_time(p[:end])} (#{p[:emoji]} #{p[:rating]})"
    end
  end

  puts "\n⭐ Best Fishing Rating: #{ratings[rating-1]} (#{rating_emoji[rating-1]})"

  # Sunrise/sunset (simplified)
  day_of_year = dt.yday
  declination = 23.44 * Math.sin((284 + day_of_year) * 360 * Math::PI / 180 / 365)
  lat_rad = lat * Math::PI / 180
  dec_rad = declination * Math::PI / 180
  cos_ha = -Math.tan(lat_rad) * Math.tan(dec_rad)
  ha = if cos_ha < -1
         Math::PI
       elsif cos_ha > 1
         0
       else
         Math.acos(cos_ha)
       end
  day_length = ha * 2 / (Math::PI / 12)
  noon = 12.0 - lon / 15.0
  sunrise = noon - day_length / 2
  sunset = noon + day_length / 2
  puts "🌅 Sunrise: #{format_time(sunrise)} | 🌇 Sunset: #{format_time(sunset)}"

  moon_rise = (solunar_data[:moon_transit] + 6) % 24
  moon_set = (solunar_data[:moon_transit] + 18) % 24
  puts "🌙 Moonrise: #{format_time(moon_rise)} | 🌇 Moonset: #{format_time(moon_set)}"
end

options = {}
OptionParser.new do |opts|
  opts.banner = "Usage: lunar_fishing.rb [options]"
  opts.on("--date DATE", "YYYY-MM-DD") { |v| options[:date] = v }
  opts.on("--lat LAT", Float, "Latitude (positive North)") { |v| options[:lat] = v }
  opts.on("--lon LON", Float, "Longitude (positive East)") { |v| options[:lon] = v }
  opts.on("--phase-only", "Output only phase name") { options[:phase_only] = true }
  opts.on("--save-location NAME", "Save location with name") { |v| options[:save_location] = v }
  opts.on("--use-location NAME", "Use saved location") { |v| options[:use_location] = v }
  opts.on("--list-locations", "List saved locations") { options[:list_locations] = true }
end.parse!

config = { "locations" => {} }
if File.exist?(CONFIG_FILE)
  config = JSON.parse(File.read(CONFIG_FILE)) rescue config
end

if options[:list_locations]
  if config["locations"].empty?
    puts "No saved locations."
    exit
  end
  puts "\n📍 Saved Locations:"
  config["locations"].each do |name, loc|
    lat_str = "#{loc['lat'].abs.round(2)}°#{loc['lat'] >= 0 ? 'N' : 'S'}"
    lon_str = "#{loc['lon'].abs.round(2)}°#{loc['lon'] >= 0 ? 'E' : 'W'}"
    puts "  #{name}: #{lat_str}, #{lon_str}"
  end
  exit
end

if options[:save_location]
  config["locations"][options[:save_location]] = { "lat" => options[:lat] || DEFAULT_LAT, "lon" => options[:lon] || DEFAULT_LON }
  File.write(CONFIG_FILE, JSON.pretty_generate(config))
  puts "✅ Location '#{options[:save_location]}' saved."
end

lat = options[:lat] || DEFAULT_LAT
lon = options[:lon] || DEFAULT_LON

if options[:use_location]
  if config["locations"][options[:use_location]]
    lat = config["locations"][options[:use_location]]["lat"]
    lon = config["locations"][options[:use_location]]["lon"]
    puts "📍 Using saved location: #{options[:use_location]}"
  else
    puts "Location '#{options[:use_location]}' not found."
    exit 1
  end
end

dt = DateTime.now
if options[:date]
  dt = DateTime.parse(options[:date])
end

render(dt, lat, lon, options[:phase_only])
