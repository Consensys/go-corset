;;error:4:9-11:symbol eq already declared
;; Cannot overload pure with impure, and vice versa.
(defpurefun (eq (x :binary) (y :binary)) (- x y))
(defun (eq x y) (+ x y))
