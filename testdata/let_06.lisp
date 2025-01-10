(defpurefun ((vanishes! :@loob) x) x)
(defcolumns X Y Z)

(defconstraint c1 ()
  (let ((XY (* X Y)))
    ;; X*Y == Z || X*Y == Z + 1
    (vanishes! (* (- XY Z) (- XY Z 1)))))
