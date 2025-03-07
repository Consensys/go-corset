(defpurefun (fn (x :binary) (y :binary)) (- x y))
(defpurefun (fn x (y :binary) z) (+ x y))
(defpurefun (fn x y a b) (* x y))

(defcolumns (X :binary@loob) (Y :binary@loob) (A :i16@loob) (B :i16@loob))
(defconstraint c1 () (fn X Y))
(defconstraint c2 () (fn A B 0 0))
