
for i in {1..8}
do
    URL="https://www.strava.com/athlete/training_activities?keywords=&sport_type=&tags=&commute=&private_activities=&trainer=&gear=&search_session_id=71db6606-5be3-4c3f-a236-3e82030ef4ef&new_activity_only=false&page=$i"

    curl $URL \
        -H 'accept: text/javascript, application/javascript, application/ecmascript, application/x-ecmascript' \
        -H 'accept-language: en-US,en;q=0.9,ru-RU;q=0.8,ru;q=0.7' \
        -H 'priority: u=1, i' \
        -H 'referer: https://www.strava.com/athlete/training' \
        -H 'sec-ch-ua: "Not(A:Brand";v="99", "Google Chrome";v="133", "Chromium";v="133"' \
        -H 'sec-ch-ua-mobile: ?0' \
        -H 'sec-ch-ua-platform: "Linux"' \
        -H 'sec-fetch-dest: empty' \
        -H 'sec-fetch-mode: cors' \
        -H 'sec-fetch-site: same-origin' \
        -H 'user-agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36' \
        -H 'x-requested-with: XMLHttpRequest' \
        > "result_$i.json"
        sleep 1000
done
