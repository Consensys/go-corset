;;error:6:26-30:not permitted in pure context
;;error:7:24-30:expected bool, found int
(defcolumns (A :i16))
(defun (getA) A)
;; not pure!
(defpurefun (fd x) (+ x (getA)))
(defconstraint test () (fd 1))
