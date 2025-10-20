;;
(defcolumns (A :i16))
(defpurefun (Id x) x)
(defconstraint test () (== 0 (Id A)))
