(defcolumns A)
(defun (getA) A)
;; not pure!
(defpurefun (id x) (+ x (getA)))
(defconstraint test () (id 1))
