;;error:5:26-27:expected type u1 (found 𝔽@loob)
(defpurefun (fn (x :binary) y) (- x y))

(defcolumns (X :@loob) (Y :@loob) (A :binary@loob) (B :binary@loob))
(defconstraint c1 () (fn X Y))
