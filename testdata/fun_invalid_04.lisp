(defcolumns X)
;; recursive :)
(defun (id x) (+ x (id x)))
;; infinite loop?
(defconstraint c1 () (id X))
