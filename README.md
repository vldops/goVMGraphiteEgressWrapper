# goVMGraphiteEgressWrapper
- go build -mod vendor -o goVMGraphiteEgressWrapper
- ./goVMGraphiteEgressWrapper --config config.yml

works:
-  graphite{target=~"victoria.metric.*"}
-  graphite{target="victoria.metric1"}
-  graphite{target="victoria.metric2"}
-  graphite{target="victoria.metric3"}
-  graphite{target="victoria.metric[1234]"}
-  graphite{target="victoria.metric{1,2,3,GGGG}"}
-  graphite{target="victoria.metric*"}
