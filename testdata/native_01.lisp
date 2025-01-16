(defpurefun ((eq! :@loob) x y) (- x y))

(defcolumns P X Y)
(defcomputed (Z) (fwd-changes-within P X))
(defconstraint c1 () (eq! Y Z))
