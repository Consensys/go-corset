(defpurefun ((vanishes! :𝔽@loob) x) x)
(defcolumns (X :i16) (Y :i16) (Z :i16))

(defconstraint c1 ()
  (let ((XY (* X Y)))
    ;; X*Y == Z || X*Y == Z + 1
    (vanishes! (* (- XY Z) (- XY Z 1)))))
