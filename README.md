### Newsmaker

Newsmaker is a daemon that implements the news filtering pipeline. The pipeline connects multiple news sources (rss/atom feeds) with multiple notifiers aka "publishers"  through multiple filters.

Right now only "http-get" notifiers are supported.  This permits you to publish the filtered news in your private Telegram channels using your own custom bot.
It is not difficult to add your own custom notifiers/publishers or even custom sources too if rss/atom is not enough.


### How to run
```
newsmaker config.toml
```

Config.toml sample:

```toml
rotation_tick = "45s" # random source will be requested each tick.
mute_hours = [20, 5] # demon will stop sources rotation and be mute from 8pm till 5 am

[[filters]] 
cond = "ABC; DAP" # title must contain either ABC _OR_ DAP
sources = ["main"] # sources to filter
pubs = ["main"]  # publishers that receive message, if cond is true

[[filters]]
cond = "Paris & Hilton"  # title must contain both Paris _AND_ Hilton (in any order)
pubs = ["info"]
sources = ["main", "other"]

[src.main]
cd = "15m" # cd is the cooldown for which the source is excluded from "rotation" after it was requested.
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

### Filter language description

Filter expression is DNF of regex pattern sequences. Before checking against the filter expression, the sentence (news title) is tokenized into a words sequence.

Filter Grammar EBNF:
```
Expr := Conj {";" Conj} // ; is OR
Conj := Seq {"&" Seq} // & is AND
Seq := Pattern {" " Pattern} //  a sequence of patterns to match some subsequence of words in a sentence.
```

Pattern is Go regex with minor *simplifications:
- Lowercase letter matches both lowercase and uppercase, but uppercase matches only uppercase
- Prefix match by default. If you need "middle" match, start pattern with star. If you need precise word match, end pattern with dollar. If youn need strict suffix match, start pattern with star and end it with dollar.

Patterns are word patterns i.e. they are matched against individual words of a sentence. If you need to
match more than 1 word, you must use a sequence of patterns `Seq`.

There is special meta-characters for Russian language nouns morphology (see words.go)


### todo

- add more meaningfull unit tests & CI
- vendoring
- better docs






 