;;error:3:14-16:symbol fn already declared
(defpurefun (fn (x :binary) (y :binary)) (- x y))
(defpurefun (fn x y) (* x y))

(defcolumns (X :binary@loob) (Y :binary@loob) (A :i16@loob) (B :i16@loob))
(defconstraint c1 () (fn X Y))
(defconstraint c2 () (fn A B))
