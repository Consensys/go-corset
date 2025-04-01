(defcolumns (A :i16))
(defpurefun ((get :i16) (e :i16)) e)
(defconstraint test () (== 0 (get A)))
