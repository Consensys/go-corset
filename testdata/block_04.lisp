(defcolumns (X :i16@loob) (Y :i16@loob))

(defconstraint c1 ()
  (begin (- X Y) (- Y X)))
