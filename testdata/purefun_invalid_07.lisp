;;error:2:1-2:blah
(defcolumns A)
(defun (getA) A)
;; not pure!
(defpurefun (id x) (+ x (getA)))
(defconstraint test () (id 1))
