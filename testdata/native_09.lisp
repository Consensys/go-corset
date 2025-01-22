(defpurefun ((eq! :@loob) x y) (- x y))

(defcolumns (P :binary) K1a K1b V1 K2a K2b R)
(defcomputed (V2) (map-if P K2a K2b P K1a K1b V1))
(defconstraint c1 () (eq! V2 R))
