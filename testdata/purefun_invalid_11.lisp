;; Duplicate overload is always a syntax error.
(defpurefun (eq (x :binary) (y :binary)) (- x y))
(defpurefun (eq (x :binary) (y :binary)) (+ x y))
