(defcolumns X Y Z)
(defconstraint c1 () (* Z (- Z 1)))
(defconstraint c2 () (* (- Y Z) (- Y Z 2)))
(defconstraint c3 () (* (- X Y) (- X Y 4)))
