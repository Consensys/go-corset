;;error:5:8-11:invalid index variable
(defcolumns X)

(defconstraint X_t1 ()
  (for (i) [1:3] (not_eq! X i)))
