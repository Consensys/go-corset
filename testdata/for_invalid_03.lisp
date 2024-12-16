(defcolumns X)

(defconstraint X_t1 ()
  (for i [] (not_eq! X i)))

(defconstraint X_t2 ()
  (for i [1:] (not_eq! X i)))

(defconstraint X_t3 ()
  (for i [:1] (not_eq! X i)))

(defconstraint X_t4 ()
  (for i [1:1:2] (not_eq! X i)))
