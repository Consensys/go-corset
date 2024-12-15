;;error:2:1-2:blah
(defcolumns A)
;; not pure!
(defpurefun (id x) (+ x A))
(defconstraint test () (id 1))
