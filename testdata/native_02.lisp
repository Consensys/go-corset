(defpurefun ((eq! :@loob) x y) (- x y))

(defcolumns P X1 X2 Y)
(defcomputed (Z) (fwd-changes-within P X1 X2))
(defconstraint c1 () (eq! Y Z))
