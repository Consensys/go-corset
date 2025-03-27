;;error:4:9-11:symbol id already declared
(defcolumns (X :i16))
;; recursive :)
(defun (id x) (+ x (id x)))
;; infinite loop?
(defconstraint c1 () (id X))
