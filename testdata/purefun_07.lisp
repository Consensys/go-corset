(defpurefun ((fn :bool) (x :binary) (y :binary)) (== x y))
(defpurefun ((fn :bool) x y z) (== 0 (* x y)))

(defcolumns (X :binary) (Y :binary) (A :i16) (B :i16))
(defconstraint c1 () (fn X Y))
(defconstraint c2 () (fn A B 0))
