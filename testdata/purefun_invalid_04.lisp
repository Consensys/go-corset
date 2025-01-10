;;error:5:25-26:not permitted in pure context
;;error:6:24-30:expected loobean constraint (found 𝔽)
(defcolumns A)
;; not pure!
(defpurefun (id x) (+ x A))
(defconstraint test () (id 1))
