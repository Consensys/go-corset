(defpurefun ((vanishes! :𝔽@loob) x) x)
;;
(defcolumns (A :i16))
(defpurefun (Id x) x)
(defconstraint test () (vanishes! (Id A)))
