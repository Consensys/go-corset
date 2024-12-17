(defcolumns (X :@loob) (Y :@loob))

(defconstraint c1 ()
  (begin (- X Y) (- Y X)))
