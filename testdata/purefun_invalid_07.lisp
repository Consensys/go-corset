;;error:6:25-31:not permitted in pure context
;;error:7:24-30:expected loobean constraint (found 𝔽)
(defcolumns A)
(defun (getA) A)
;; not pure!
(defpurefun (id x) (+ x (getA)))
(defconstraint test () (id 1))
