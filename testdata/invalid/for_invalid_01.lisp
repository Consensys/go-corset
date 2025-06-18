;;error:5:3-16:expected 3 arguments, found 2
(defcolumns (X :i16))
;; X != 1 && X != 2 && X != 3
(defconstraint X_t1 ()
  (for i [1:3]))
