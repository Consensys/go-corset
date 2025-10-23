(defcolumns (X :i16) (Y :i16))
(defconstraint c1 () (== 0 (* Y (- Y 1) (- Y 2) (- Y 3))))
(defconstraint c2 () (== 0 (* (- X Y) (- X Y 4))))
;;
(defconstraint c3 () (∨ (== 0 Y) (== Y 1) (== Y 2) (== Y 3)))
(defconstraint c4 () (∨ (== X Y) (== X (+ Y 4))))
