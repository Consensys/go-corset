(defcolumns X Y)
(defconstraint c1 () (* Y (- Y 1) (- Y 2) (- Y 3)))
(defconstraint c2 () (* (- X Y) (- X Y 4)))
