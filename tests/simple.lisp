(def test (n)
     (if (< n 2)
	 (print 5)
       (print 6)))

(def main ()
     (test (- 1 (test 1))))
