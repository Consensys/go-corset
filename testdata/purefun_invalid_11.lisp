;;error:4:1-50:symbol eq already declared
;; Duplicate overload is always a syntax error.
(defpurefun (eq (x :binary) (y :binary)) (- x y))
(defpurefun (eq (x :binary) (y :binary)) (+ x y))
