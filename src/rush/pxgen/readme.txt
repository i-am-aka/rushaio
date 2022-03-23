task opens WS session with pxapi
  api opens pxgenjs with tmpdir unix pipe for http proxying

  pxgenjs fetches latest px js, wraps, execs

  px js issues first collector request. pxgenjs captures and sends over pipe

  task recv request from ws

  task execute request

  task send response to ws

  ws send px3 to task, conn close

task use px3

...

PROFIT
