;;error:5:26-27:expected type u1 (found u16)
(defpurefun (fn (x :binary) y) (- x y))

(defcolumns (X :i16) (Y :i16) (A :binary) (B :binary))
(defconstraint c1 () (fn X Y))
