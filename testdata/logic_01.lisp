(defcolumns (X :i16) (Y :i16))

(defconstraint c1 ()
  (or! (eq! X 0) (eq! Y 0)))

(defconstraint c2 ()
  (and! (neq! X 1) (neq! Y 1)))

(defconstraint c3 ()
  (and! (not! (eq! X 1)) (neq! Y 1)))

(defconstraint c4 ()
  (and! (neq! X 1) (not! (eq! Y 1))))
