(defpurefun (vanishes! x) (== 0 x))
;;
(defcolumns (A :i16))
(defpurefun (Id x) x)
(defconstraint test () (vanishes! (Id A)))
