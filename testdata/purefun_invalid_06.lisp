;;error:2:1-2:blah
(defcolumns (X :i16))
;; recursive :)
(defpurefun (id x) (+ x (id x)))
;; infinite loop?
(defconstraint c1 () (id X))
