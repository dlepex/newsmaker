### newsmaker

Newsmaker is a demon who implements the simple pipeline, that connects multiple news sources (rss/atom feeds) with multiple notifiers aka "publishers"  thru multiple filters.

Right now only "http GET" notifiers are supported.  It, for instance, permits you to publish the filtered news in Telegram channels using your own custom bot.
Its very easy to add your own custom notifiers, and custom sources (if atom/rss is not enough), as they have simple interface.


##### How to run 
```
newsmaker config.toml
```

Config.toml sample:

```toml
rotation_tick = "45s"
mute_hours = [20, 5] # demon will be silent from 20pm till 5 am

[[filters]]
cond = "P(\x26)G; DAP;"
sources = ["main"]
pubs = ["main"]

[[filters]]
cond = "Paris & Hilton" 
pubs = ["info"]
sources = ["main", "other"]

[src.main]
cd = "15m"
links = ["https://regnum.ru/rss/polit", "https://regnum.ru/rss/accidents"]

[src.other]
cd = "15m"
links = ["https://news.yandex.ru/finances.rss"]

[pub.main]
get_url = "https://api.telegram.org/bot50034962:BBGuVfL-EZ-Wnlj1b80oysOkurJgZdbI/sendMessage?text=%s&chat_id=-20023152348394761&parse_mode=Markdown"

[pub.info]
send_pause = "5s"
get_url = "https://api.telegram.org/bot50034962:BBGuVfL-EZ-Wnlj1b80oysOkurJgZdbI/sendMessage?text=%s&chat_id=-20023152348394761&parse_mode=Markdown"
```





###### todo

- add more meaningfull unit tests & CI
- vendoring







 