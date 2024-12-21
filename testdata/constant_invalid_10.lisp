;;error:12:13-14:symbol X already declared
;;error:13:22-37:expected loobean constraint (found 𝔽)
;;error:14:22-45:expected loobean constraint (found 𝔽)
;;error:15:22-46:expected loobean constraint (found 𝔽)
(defconst
  X     1
  ONE   X
  TWO   (+ 1 ONE)
  FOUR  (* 2 TWO)
)

(defcolumns X Y Z)
(defconstraint c1 () (* Z (- Z ONE)))
(defconstraint c2 () (* (- Y Z) (- Y Z TWO)))
(defconstraint c3 () (* (- X Y) (- X Y FOUR)))
