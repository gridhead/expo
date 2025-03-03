package base

const Headtemp = "[SN#%d] %s"

const Bodytemp = `
%s

_This issue ticket was originally created [here](%s) on a Pagure repository, [**%s**](https://%s/%s) by [%s](%s) on **%s**._

_This issue ticket was automatically created by [**Pagure Exporter**](https://github.com/gridhead/pagure-exporter)._
`

const Chattemp = `
%s

_This comment was originally created [here](%s#comment-%d) by [**%s**](%s) under [this](%s) issue ticket on a Pagure repository, [**%s**](https://%s/%s) on **%s**._

_This comment was automatically created by [**Pagure Exporter**](https://github.com/gridhead/pagure-exporter)._
`
