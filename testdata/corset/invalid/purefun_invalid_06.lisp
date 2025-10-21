;;error:4:28-37:recursion not permitted here
(defcolumns (X :i16))
;; recursive :)
(defpurefun (recfn x) (+ x (recfn x)))
;; infinite loop?
(defconstraint c1 () (== 0 (recfn X)))
