(defcolumns (X :i16) (Y :i16) (Z :i16))

(defconstraint c1 ()
  (let ((XY (* X Y)))
    ;; X*Y == Z || X*Y == Z + 1
    (== 0 (* (- XY Z) (- XY Z 1)))))
