(defcolumns (X :i16) (Y :i16))

(defconstraint c1 ()
  (begin (== X Y) (== Y X)))
