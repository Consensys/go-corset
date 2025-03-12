;;error:6:25-31:not permitted in pure context
;;error:7:24-30:expected loobean constraint (found u16)
(defcolumns (A :i16))
(defun (getA) A)
;; not pure!
(defpurefun (fd x) (+ x (getA)))
(defconstraint test () (fd 1))
