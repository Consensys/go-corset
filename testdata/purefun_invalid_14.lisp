;;error:3:14-16:symbol fn already declared
(defpurefun (fn (x :binary) (y :binary)) (- x y))
(defpurefun (fn x y) (* x y))

(defcolumns (X :binary) (Y :binary) (A :i16) (B :i16))
(defconstraint c1 () (fn X Y))
(defconstraint c2 () (fn A B))
