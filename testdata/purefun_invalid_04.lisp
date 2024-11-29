(defcolumns A)
;; not pure!
(defpurefun (id x) (+ x A))
(defconstraint test () (id 1))
