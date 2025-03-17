(defpurefun (fn (x :binary) (y :binary)) (- x y))
(defpurefun (fn x (y :binary) z) (+ x y))
(defpurefun (fn x y a b) (* x y))

(defcolumns (X :binary) (Y :binary) (A :i16) (B :i16))
(defconstraint c1 () (fn X Y))
(defconstraint c2 () (fn A B 0 0))
