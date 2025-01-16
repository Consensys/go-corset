(defpurefun ((eq! :@loob) x y) (- x y))

(defcolumns P Q X Y)
(defcomputed (Z) (bwd-fill-within P Q X))
(defconstraint c1 () (eq! Y Z))
