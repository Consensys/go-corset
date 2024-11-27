(defcolumns X)
;; recursive :)
(defpurefun (id x) (+ x (id x)))
;; infinite loop?
(defconstraint c1 () (id X))
