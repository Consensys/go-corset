;;error:5:25-31:not permitted in pure context
(defcolumns A)
(defun (getA) A)
;; not pure!
(defpurefun (id x) (+ x (getA)))
(defconstraint test () (id 1))
