;;error:8:23-25:ambiguous invocation
(defpurefun (fn (x :binary) y) (- x y))
(defpurefun (fn x (y :binary)) (+ x y))
(defpurefun (fn x y) (* x y))

(defcolumns (X :@loob) (Y :@loob) (A :binary@loob) (B :binary@loob))
(defconstraint c1 () (fn X Y)) ;; not ambiguous
(defconstraint c2 () (fn A B)) ;; ambiguous
