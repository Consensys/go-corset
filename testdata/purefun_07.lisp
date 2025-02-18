(defpurefun (fn (x :binary) (y :binary)) (- x y))
(defpurefun (fn x y z) (* x y))

(defcolumns (X :binary@loob) (Y :binary@loob) (A :@loob) (B :@loob))
(defconstraint c1 () (fn X Y))
(defconstraint c2 () (fn A B 0))
