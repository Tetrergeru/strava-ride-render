for map in $(cat maps.txt); do
  curl "https://www.strava.com/activities/$map/streams?stream_types[]=time&stream_types[]=watts_calc&stream_types[]=altitude&stream_types[]=heartrate&stream_types[]=cadence&stream_types[]=temp&stream_types[]=distance&stream_types[]=grade_smooth&stream_types[]=latlng&_=1740694610814" \
    -H 'accept: text/javascript, application/javascript, application/ecmascript, application/x-ecmascript' \
    -H 'accept-language: en-US,en;q=0.9,ru-RU;q=0.8,ru;q=0.7' \
    -H 'priority: u=1, i' \
    -H 'sec-ch-ua: "Not(A:Brand";v="99", "Google Chrome";v="133", "Chromium";v="133"' \
    -H 'sec-ch-ua-mobile: ?0' \
    -H 'sec-ch-ua-platform: "Linux"' \
    -H 'sec-fetch-dest: empty' \
    -H 'sec-fetch-mode: cors' \
    -H 'sec-fetch-site: same-origin' \
    -H 'user-agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36' \
    -H 'x-requested-with: XMLHttpRequest' \
    > "./$map.json"

  sleep 1s
done
