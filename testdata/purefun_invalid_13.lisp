;;error:5:22-30:expected bool, found int
(defpurefun (fn (x :binary) y) (- x y))

(defcolumns (X :i16) (Y :i16) (A :binary) (B :binary))
(defconstraint c1 () (fn X Y))
