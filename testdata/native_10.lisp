(defpurefun ((eq! :@loob) x y) (- x y))

(defcolumns P Q (X :i8) Y)
(defcomputed (Z) (fwd-fill-within P Q X))
(defconstraint c1 () (eq! Y Z))
