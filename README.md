### newsmaker

Newsmaker is a demon that implements the news filtering pipeline. The pipeline connects multiple news sources (rss/atom feeds) with multiple notifiers aka "publishers"  thru multiple filters.

Right now only "http GET" notifiers are supported.  It, for instance, permits you to publish the filtered news in your private Telegram channels using your own custom bot.
Its very easy to add your own custom notifiers andcustom sources too if rss/atom is not enough.


##### How to run
```
newsmaker config.toml
```

Config.toml sample:

```toml
rotation_tick = "45s"
mute_hours = [20, 5] # demon will be silent from 20pm till 5 am

[[filters]]  
cond = "P(\x26)G; DAP;" # title must contain either P&G _OR_ DAP
sources = ["main"]
pubs = ["main"] 

[[filters]]
cond = "Paris & Hilton"  # title must contain both Paris _AND_ Hilton (in any order)
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

###### Filter language description

Filter expression is a DNF of regex pattern sequences. Before checking agaisnt the filter expression, the sentence (news title) is tokenized into the words sequence.

Grammar EBF:
```
Expr := Conj {";" Conj} // ; is OR
Conj := Seq {"&" Seq} // & is AND
Seq := Pattern {" " Pattern} //  Seq a sequence of patterns that matches the subsequence of words in the sentence.
```

Pattern is basically a normal golang regex with minor simplifications. It also has special meta-characters for Russian language nouns morphology. (see words.go code)
Patterns are word patterns i.e. they are applied to separate words, not to the sentence as a whole.

*Pattern simplications:*
- Prefix match by default. If you need suffix match, start pattern with star. If you need precise match, end pattern with dollar.
- Lowercase letter matches both lowercase and uppercase, but uppercase matches only uppercase



###### Todo

- add more meaningfull unit tests & CI
- vendoring
- better docs






 