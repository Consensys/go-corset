;;error:5:24-25:not permitted in pure context
;;error:6:24-29:expected bool, found int
(defcolumns (A :i16))
;; not pure!
(defpurefun (f x) (+ x A))
(defconstraint test () (f 1))
