;;error:4:25-26:not permitted in pure context
(defcolumns A)
;; not pure!
(defpurefun (id x) (+ x A))
(defconstraint test () (id 1))
