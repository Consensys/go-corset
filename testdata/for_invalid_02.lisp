(defcolumns X)

(defconstraint X_t1 ()
  (for (i) [1:3] (not_eq! X i)))
