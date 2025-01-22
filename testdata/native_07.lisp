(defpurefun ((eq! :@loob) x y) (- x y))

(defcolumns (P :binary) K1 V1 K2 R)
(defcomputed (V2) (map-if P K2 P K1 V1))
(defconstraint c1 () (eq! V2 R))
