(defconst
  ONE_  1
  ONE   ONE_
  TWO   (+ 1 ONE)
  FOUR  (* 2 TWO)
)

(defcolumns X Y Z)
(defconstraint c1 () (* Z (- Z ONE)))
(defconstraint c2 () (* (- Y Z) (- Y Z TWO)))
(defconstraint c3 () (* (- X Y) (- X Y FOUR)))
