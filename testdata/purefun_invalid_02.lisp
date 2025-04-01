;;error:4:28-30:unknown symbol
(defcolumns (A :i16))
(defpurefun (dd x) x)
(defconstraint test () (== dd A))
